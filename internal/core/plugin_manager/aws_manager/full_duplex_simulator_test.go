package aws_manager

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func server(timeout time.Duration) (string, func(), error) {
	routine.InitPool(1024)

	port, err := network.GetRandomPort()
	if err != nil {
		return "", nil, err
	}

	data := map[string]chan []byte{}
	data_mu := sync.Mutex{}

	eng := gin.New()
	eng.POST("/invoke", func(c *gin.Context) {
		id := c.Request.Header.Get("x-dify-plugin-request-id")
		var ch chan []byte

		data_mu.Lock()
		if _, ok := data[id]; !ok {
			ch = make(chan []byte, 1024)
			data[id] = ch
		} else {
			ch = data[id]
		}
		data_mu.Unlock()

		time.AfterFunc(timeout, func() {
			c.Request.Body.Close()
		})

		// read data asynchronously
		for {
			buf := make([]byte, 1024)
			n, err := c.Request.Body.Read(buf)
			if err != nil {
				break
			}
			ch <- buf[:n]
		}

		c.Status(http.StatusOK)
	})

	eng.GET("/response", func(ctx *gin.Context) {
		id := ctx.Request.Header.Get("x-dify-plugin-request-id")
		var ch chan []byte
		data_mu.Lock()
		if _, ok := data[id]; ok {
			ch = data[id]
		} else {
			ch = make(chan []byte, 1024)
			data[id] = ch
		}
		data_mu.Unlock()

		ctx.Writer.WriteHeader(http.StatusOK)
		ctx.Writer.Header().Set("Content-Type", "application/octet-stream")
		ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
		ctx.Writer.Header().Set("Connection", "keep-alive")
		ctx.Writer.Write([]byte("pong\n"))
		ctx.Writer.Flush()

		for {
			select {
			case data := <-ch:
				ctx.Writer.Write(data)
				ctx.Writer.Flush()
			case <-ctx.Done():
				return
			case <-ctx.Writer.CloseNotify():
				return
			case <-time.After(timeout):
				ctx.Status(http.StatusOK)
				return
			}
		}
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: eng,
	}

	go func() {
		srv.ListenAndServe()
	}()

	return fmt.Sprintf("http://localhost:%d", port), func() {
		srv.Close()
	}, nil
}

func TestFullDuplexSimulator_Send(t *testing.T) {
	url, cleanup, err := server(time.Second * 100)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	time.Sleep(time.Second)

	simulator, err := NewFullDuplexSimulator(url)
	if err != nil {
		t.Fatal(err)
	}

	recved := make([]byte, 0)

	simulator.On(func(data []byte) {
		if len(bytes.TrimSpace(data)) == 0 {
			return
		}

		recved = append(recved, data...)
	})

	if done, err := simulator.StartTransaction(); err != nil {
		t.Fatal(err)
	} else {
		defer done()
	}

	if err := simulator.Send([]byte("hello\n")); err != nil {
		t.Fatal(err)
	}
	if err := simulator.Send([]byte("world\n")); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 500)

	if string(recved) != "hello\nworld\n" {
		t.Fatal(fmt.Sprintf("recved: %s", string(recved)))
	}
}
