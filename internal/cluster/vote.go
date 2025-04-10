package cluster

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/cache"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/http_requests"
)

func (c *Cluster) voteAddresses() error {
	c.notifyVoting()
	defer c.notifyVotingCompleted()
	var totalErrors error
	addError := func(err error) {
		if err != nil {
			if totalErrors == nil {
				totalErrors = err
			} else {
				totalErrors = errors.Join(totalErrors, err)
			}
		}
	}

	// get all nodes status
	nodes, err := cache.GetMap[node](CLUSTER_STATUS_HASH_MAP_KEY)
	if err == cache.ErrNotFound {
		return nil
	}

	for node_id, nodeStatus := range nodes {
		if node_id == c.id {
			continue
		}

		// vote for ips
		ipsVoting := make(map[string]bool)
		for _, addr := range nodeStatus.Addresses {
			// skip ips which have already been voted by current node in the last 5 minutes
			for _, vote := range addr.Votes {
				if vote.NodeID == c.id {
					if time.Since(time.Unix(vote.VotedAt, 0)) < time.Minute*5 && !vote.Failed {
						continue
					} else if time.Since(time.Unix(vote.VotedAt, 0)) < time.Minute*30 && vote.Failed {
						continue
					}
				}
			}

			ipsVoting[addr.fullAddress()] = time.Since(time.Unix(nodeStatus.LastPingAt, 0)) < c.nodeDisconnectedTimeout
		}

		// lock the node status
		if err := c.LockNodeStatus(node_id); err != nil {
			addError(err)
			c.UnlockNodeStatus(node_id)
			continue
		}

		// get the node status again
		nodeStatus, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, node_id)
		if err != nil {
			addError(err)
			c.UnlockNodeStatus(node_id)
			continue
		}

		// update the node status
		for i, ip := range nodeStatus.Addresses {
			// update voting time
			if success, ok := ipsVoting[ip.fullAddress()]; ok {
				// check if the ip has already voted
				alreadyVoted := false
				for j, vote := range ip.Votes {
					if vote.NodeID == c.id {
						nodeStatus.Addresses[i].Votes[j].VotedAt = time.Now().Unix()
						nodeStatus.Addresses[i].Votes[j].Failed = !success
						alreadyVoted = true
						break
					}
				}
				// add a new vote
				if !alreadyVoted {
					nodeStatus.Addresses[i].Votes = append(nodeStatus.Addresses[i].Votes, vote{
						NodeID:  c.id,
						VotedAt: time.Now().Unix(),
						Failed:  !success,
					})
				}
			}
		}

		// sync the node status
		if err := cache.SetMapOneField(CLUSTER_STATUS_HASH_MAP_KEY, node_id, nodeStatus); err != nil {
			addError(err)
		}

		// unlock the node status
		if err := c.UnlockNodeStatus(node_id); err != nil {
			addError(err)
		}
	}

	return totalErrors
}

func (c *Cluster) voteAddress(addr address) error {
	type healthcheck struct {
		Status string `json:"status"`
	}

	healthcheckEndpoint, err := url.JoinPath(fmt.Sprintf("http://%s:%d", addr.Ip, addr.Port), "health/check")
	if err != nil {
		return err
	}

	resp, err := http_requests.GetAndParse[healthcheck](
		http.DefaultClient,
		healthcheckEndpoint,
		http_requests.HttpWriteTimeout(500),
		http_requests.HttpReadTimeout(500),
	)

	if err != nil {
		return err
	}

	if resp.Status != "ok" {
		return errors.New("health check failed")
	}

	return nil
}

func (c *Cluster) SortIps(nodeStatus node) []address {
	// sort by votes
	sort.Slice(nodeStatus.Addresses, func(i, j int) bool {
		return len(nodeStatus.Addresses[i].Votes) > len(nodeStatus.Addresses[j].Votes)
	})

	return nodeStatus.Addresses
}
