package network

import "net"

func GetRandomPort() (uint16, error) {
	// generate a random port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	listener.Close()

	port := listener.Addr().(*net.TCPAddr).Port
	return uint16(port), nil
}
