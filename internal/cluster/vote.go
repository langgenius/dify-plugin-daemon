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

func (c *Cluster) voteIps() error {
	c.notifyVoting()
	defer c.notifyVotingCompleted()
	var total_errors error
	add_error := func(err error) {
		if err != nil {
			if total_errors == nil {
				total_errors = err
			} else {
				total_errors = errors.Join(total_errors, err)
			}
		}
	}

	// get all nodes status
	nodes, err := cache.GetMap[node](CLUSTER_STATUS_HASH_MAP_KEY)
	if err == cache.ErrNotFound {
		return nil
	}

	for node_id, node_status := range nodes {
		if node_id == c.id {
			continue
		}

		// vote for ips
		ips_voting := make(map[string]bool)
		for _, ip := range node_status.Ips {
			// skip ips which have already been voted by current node in the last 5 minutes
			for _, vote := range ip.Votes {
				if vote.NodeID == c.id {
					if time.Since(time.Unix(vote.VotedAt, 0)) < time.Minute*5 && !vote.Failed {
						continue
					} else if time.Since(time.Unix(vote.VotedAt, 0)) < time.Minute*30 && vote.Failed {
						continue
					}
				}
			}

			ips_voting[ip.Address] = c.voteIp(ip) == nil
		}

		// lock the node status
		if err := c.LockNodeStatus(node_id); err != nil {
			add_error(err)
			c.UnlockNodeStatus(node_id)
			continue
		}

		// get the node status again
		node_status, err := cache.GetMapField[node](CLUSTER_STATUS_HASH_MAP_KEY, node_id)
		if err != nil {
			add_error(err)
			c.UnlockNodeStatus(node_id)
			continue
		}

		// update the node status
		for i, ip := range node_status.Ips {
			// update voting time
			if success, ok := ips_voting[ip.Address]; ok {
				// check if the ip has already voted
				already_voted := false
				for j, vote := range ip.Votes {
					if vote.NodeID == c.id {
						node_status.Ips[i].Votes[j].VotedAt = time.Now().Unix()
						node_status.Ips[i].Votes[j].Failed = !success
						already_voted = true
						break
					}
				}
				// add a new vote
				if !already_voted {
					node_status.Ips[i].Votes = append(node_status.Ips[i].Votes, vote{
						NodeID:  c.id,
						VotedAt: time.Now().Unix(),
						Failed:  !success,
					})
				}
			}
		}

		// sync the node status
		if err := cache.SetMapOneField(CLUSTER_STATUS_HASH_MAP_KEY, node_id, node_status); err != nil {
			add_error(err)
		}

		// unlock the node status
		if err := c.UnlockNodeStatus(node_id); err != nil {
			add_error(err)
		}
	}

	return total_errors
}

func (c *Cluster) voteIp(ip ip) error {
	type healthcheck struct {
		Status string `json:"status"`
	}

	healthcheck_endpoint, err := url.JoinPath(fmt.Sprintf("http://%s:%d", ip.Address, c.port), "health/check")
	if err != nil {
		return err
	}

	resp, err := http_requests.GetAndParse[healthcheck](
		http.DefaultClient,
		healthcheck_endpoint,
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

func (c *Cluster) SortIps(node_status node) []ip {
	// sort by votes
	sort.Slice(node_status.Ips, func(i, j int) bool {
		return len(node_status.Ips[i].Votes) > len(node_status.Ips[j].Votes)
	})

	return node_status.Ips
}
