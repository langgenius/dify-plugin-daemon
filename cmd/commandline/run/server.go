package run

import (
	"fmt"
	"net"
	"os"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
)

// createTCPServer creates a stream of clients that are connected to the plugin through a TCP connection
// It continuously accepts new connections and sends them to the stream
func createTCPServer(payload *RunPluginPayload) (*stream.Stream[client], error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", payload.TcpServerHost, payload.TcpServerPort))
	if err != nil {
		return nil, err
	}

	addr := listener.Addr().(*net.TCPAddr)
	payload.TcpServerHost = addr.IP.String()
	payload.TcpServerPort = addr.Port

	stream := stream.NewStream[client](30)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			stream.Write(client{
				reader: conn,
				writer: conn,
				cancel: func() {
					conn.Close()
				},
			})
		}
	}()

	return stream, nil
}

// createStdioServer creates a stream of clients that are connected to the plugin through stdin and stdout
func createStdioServer() *stream.Stream[client] {
	reader, writer := os.Stdin, os.Stdout
	stream := stream.NewStream[client](1)
	stream.Write(client{
		reader: reader,
		writer: writer,
		cancel: func() {
			reader.Close()
			writer.Close()
		},
	})

	return stream
}
