package aws_manager

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func server(recv_timeout time.Duration, send_timeout time.Duration) (string, func(), error) {
	routine.InitPool(1024)

	port, err := network.GetRandomPort()
	if err != nil {
		return "", nil, err
	}

	data := map[string]chan []byte{}
	data_mu := sync.Mutex{}

	recved := 0

	eng := gin.New()
	eng.POST("/invoke", func(c *gin.Context) {
		// fmt.Println("new send request")
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

		time.AfterFunc(send_timeout, func() {
			c.Request.Body.Close()
		})

		// read data asynchronously
		for {
			buf := make([]byte, 1024)
			n, err := c.Request.Body.Read(buf)
			if n != 0 {
				recved += n
				ch <- buf[:n]
			}
			if err != nil {
				break
			}
		}

		// output closed
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write([]byte("closed\n"))
		c.Writer.Flush()
	})

	response := 0

	eng.GET("/response", func(ctx *gin.Context) {
		// fmt.Println("new recv request")
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

		timer := time.NewTimer(recv_timeout)

		for {
			select {
			case data := <-ch:
				ctx.Writer.Write(data)
				ctx.Writer.Flush()
				response += len(data)
			case <-ctx.Done():
				return
			case <-ctx.Writer.CloseNotify():
				return
			case <-timer.C:
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
		fmt.Printf("recved: %d, responsed: %d\n", recved, response)
	}, nil
}

func TestFullDuplexSimulator_SingleSendAndReceive(t *testing.T) {
	url, cleanup, err := server(time.Second*100, time.Second*100)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	time.Sleep(time.Second)

	simulator, err := NewFullDuplexSimulator(url, time.Second*100, time.Second*100)
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

func TestFullDuplexSimulator_AutoReconnect(t *testing.T) {
	// hmmm, to ensure the server is stable, we need to run the test 100 times
	// don't ask me why, just trust me, I have spent 1 days to handle this race condition
	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			url, cleanup, err := server(time.Millisecond*700, time.Second*10)
			if err != nil {
				t.Fatal(err)
			}
			defer cleanup()

			time.Sleep(time.Second)

			simulator, err := NewFullDuplexSimulator(url, time.Millisecond*700, time.Second*10)
			if err != nil {
				t.Fatal(err)
			}

			l := 0
			recved := strings.Builder{}
			simulator.On(func(data []byte) {
				l += len(data)
				recved.Write(data)
			})

			done, err := simulator.StartTransaction()
			if err != nil {
				t.Fatal(err)
			}
			defer done()

			ticker := time.NewTicker(time.Millisecond * 1)
			counter := 0

			for range ticker.C {
				if err := simulator.Send([]byte(fmt.Sprintf("%05d", counter))); err != nil {
					t.Fatal(err)
				}
				counter++
				if counter == 3000 {
					break
				}
			}

			time.Sleep(time.Millisecond * 500)

			if l != 3000*5 {
				sent, received := simulator.GetStats()
				t.Errorf(fmt.Sprintf("expected: %d, actual: %d, sent: %d, received: %d", 3000*5, l, sent, received))
				// to find which one is missing
				for i := 0; i < 3000; i++ {
					if !strings.Contains(recved.String(), fmt.Sprintf("%05d", i)) {
						t.Errorf(fmt.Sprintf("missing: %d", i))
					}
				}
			}
		}()
	}

	wg.Wait()
}
