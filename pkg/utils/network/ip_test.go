package network

import (
	"net"
	"testing"
)

func TestIsAllIpsAvailableIPv4(t *testing.T) {
	ips, err := FetchCurrentIps()
	if err != nil {
		t.Error(err)
	}

	for _, ip := range ips {
		if net.ParseIP(ip.String()).To4() == nil {
			t.Errorf("invalid ipv4: %s", ip.String())
		}
	}
}
