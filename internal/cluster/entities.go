package cluster

import (
	"fmt"
	"time"
)

type address struct {
	Ip    string `json:"ip"`
	Port  uint16 `json:"port"`
	Votes []vote `json:"vote"`
}

func (a *address) fullAddress() string {
	return fmt.Sprintf("%s:%d", a.Ip, a.Port)
}

type vote struct {
	NodeID  string `json:"node_id"`
	VotedAt int64  `json:"voted_at"`
	Failed  bool   `json:"failed"`
}

type node struct {
	Addresses  []address `json:"ips"`
	LastPingAt int64     `json:"last_ping_at"`
}

func (c *node) available() bool {
	return time.Since(time.Unix(c.LastPingAt, 0)) < NODE_DISCONNECTED_TIMEOUT
}

type newNodeEvent struct {
	NodeID string `json:"node_id"`
}
