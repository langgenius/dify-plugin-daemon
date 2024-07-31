package cluster

import "time"

type ip struct {
	Address string `json:"address"`
	Votes   []vote `json:"vote"`
}

type vote struct {
	NodeID  string `json:"node_id"`
	VotedAt int64  `json:"voted_at"`
	Failed  bool   `json:"failed"`
}

type node struct {
	Ips        []ip  `json:"ips"`
	LastPingAt int64 `json:"last_ping_at"`
}

func (c *node) available() bool {
	return time.Since(time.Unix(c.LastPingAt, 0)) < NODE_DISCONNECTED_TIMEOUT
}
