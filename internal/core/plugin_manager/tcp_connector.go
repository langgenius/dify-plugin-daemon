package plugin_manager

import (
	"fmt"

	"github.com/panjf2000/gnet/v2"
)

type difyServer struct {
	gnet.BuiltinEventEngine

	eng       gnet.Engine
	addr      string
	multicore bool
}

func (s *difyServer) OnOpen(c gnet.Conn) (out []byte, action gnet.Action) {
	c.SetContext(&codec{})
	return nil, gnet.None
}

func (s *difyServer) OnBoot(c gnet.Engine) (action gnet.Action) {
	return gnet.None
}

func (s *difyServer) OnTraffic(c gnet.Conn) (action gnet.Action) {
	codec := c.Context().(*codec)
	messages, err := codec.Decode(c)
	if err != nil {
		return gnet.Close
	}

	for _, message := range messages {
		fmt.Println(message)
	}

	return gnet.None
}

func traffic() {
	addr := "tcp://:9000"
	multicore := true
	s := &difyServer{addr: addr, multicore: multicore}

	gnet.Run(s, addr, gnet.WithMulticore(multicore), gnet.WithNumEventLoop(8))
}
