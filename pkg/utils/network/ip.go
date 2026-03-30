package network

import "net"

// FetchCurrentIps fetches the current IP addresses of the machine
// only IPv4 addresses are returned
func FetchCurrentIps() ([]net.IP, error) {
	ips := []net.IP{}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ips, err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ips = append(ips, ipNet.IP)
			}
		}
	}

	return ips, nil
}
