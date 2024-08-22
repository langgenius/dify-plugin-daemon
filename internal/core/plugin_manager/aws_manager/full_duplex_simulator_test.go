package aws_manager

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/debugging"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/network"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
)

func init() {
	routine.InitPool(1024)
}

type S struct {
	srv  *http.Server
	url  string
	port int

	send_count          int32
	recv_buffered_count int32
	recv_count          int32

	send_request int32
	recv_request int32

	data_mu sync.Mutex
	data    map[string]chan []byte

	current_recv_request_id string
	current_send_request_id string
}

func (s *S) Stop() {
	s.srv.Close()
}

func server() (*S, error) {
	port, err := network.GetRandomPort()
	if err != nil {
		return nil, err
	}

	eng := gin.New()
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: eng,
	}

	s := &S{
		srv:  srv,
		url:  fmt.Sprintf("http://localhost:%d", port),
		data: make(map[string]chan []byte),

		send_count: 0,
		recv_count: 0,
	}

	// avoid log
	gin.SetMode(gin.ReleaseMode)

	eng.POST("/invoke", func(c *gin.Context) {
		atomic.AddInt32(&s.send_request, 1)
		defer atomic.AddInt32(&s.send_request, -1)

		id := c.Request.Header.Get("x-dify-plugin-request-id")
		max_alive_time := c.Request.Header.Get("x-dify-plugin-max-alive-time")
		s.current_send_request_id = id

		var ch chan []byte

		s.data_mu.Lock()
		if _, ok := s.data[id]; !ok {
			ch = make(chan []byte)
			s.data[id] = ch
		} else {
			ch = s.data[id]
		}
		s.data_mu.Unlock()

		timeout, err := strconv.ParseInt(max_alive_time, 10, 64)
		if err != nil {
			timeout = 60
		}

		time.AfterFunc(time.Millisecond*time.Duration(timeout), func() {
			c.Request.Body.Close()
		})

		// read data asynchronously
		for {
			buf := make([]byte, 1024)
			n, err := c.Request.Body.Read(buf)
			if n != 0 {
				atomic.AddInt32(&s.recv_buffered_count, int32(n))
				ch <- buf[:n]
				atomic.AddInt32(&s.recv_count, int32(n))
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

	eng.GET("/response", func(ctx *gin.Context) {
		atomic.AddInt32(&s.recv_request, 1)
		defer atomic.AddInt32(&s.recv_request, -1)

		// fmt.Println("new recv request")
		id := ctx.Request.Header.Get("x-dify-plugin-request-id")
		max_alive_time := ctx.Request.Header.Get("x-dify-plugin-max-alive-time")
		max_sending_bytes := ctx.Request.Header.Get("x-dify-plugin-max-sending-bytes")
		max_sending_bytes_int, err := strconv.ParseInt(max_sending_bytes, 10, 64)
		if err != nil {
			max_sending_bytes_int = 1024 * 1024
		}

		sent_bytes := int32(0)

		s.current_recv_request_id = id

		var ch chan []byte
		s.data_mu.Lock()
		if _, ok := s.data[id]; ok {
			ch = s.data[id]
		} else {
			ch = make(chan []byte)
			s.data[id] = ch
		}
		s.data_mu.Unlock()

		ctx.Writer.WriteHeader(http.StatusOK)
		ctx.Writer.Header().Set("Content-Type", "application/octet-stream")
		ctx.Writer.Header().Set("Transfer-Encoding", "chunked")
		ctx.Writer.Header().Set("Connection", "keep-alive")
		ctx.Writer.Write([]byte("pong\n"))
		ctx.Writer.Flush()

		timeout, err := strconv.ParseInt(max_alive_time, 10, 64)
		if err != nil {
			timeout = 60
		}

		timer := time.NewTimer(time.Millisecond * time.Duration(timeout))

		for {
			select {
			case data := <-ch:
				if sent_bytes+int32(len(data)) > int32(max_sending_bytes_int) {
					ctx.Writer.Write(data)
					ctx.Writer.Flush()
					ctx.Status(http.StatusOK)
					return
				}

				ctx.Writer.Write(data)
				ctx.Writer.Flush()
				atomic.AddInt32(&s.send_count, int32(len(data)))
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

	go func() {
		srv.ListenAndServe()
	}()

	return s, nil
}

func TestFullDuplexSimulator_SingleSendAndReceive(t *testing.T) {
	log.SetShowLog(false)
	defer log.SetShowLog(true)

	srv, err := server()
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	time.Sleep(time.Second)

	simulator, err := NewFullDuplexSimulator(
		srv.url, &FullDuplexSimulatorOption{
			SendingConnectionMaxAliveTime:         time.Second * 100,
			ReceivingConnectionMaxAliveTime:       time.Second * 100,
			TargetSendingConnectionMaxAliveTime:   time.Second * 99,
			TargetReceivingConnectionMaxAliveTime: time.Second * 101,
			MaxSingleRequestSendingBytes:          1024 * 1024,
			MaxSingleRequestReceivingBytes:        1024 * 1024,
		},
	)
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
	log.SetShowLog(false)
	defer log.SetShowLog(true)

	// hmmm, to ensure the server is stable, we need to run the test 100 times
	// don't ask me why, just trust me, I have spent 1 days to correctly handle this race condition
	wg := sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			srv, err := server()
			if err != nil {
				t.Fatal(err)
			}
			defer srv.Stop()

			time.Sleep(time.Second)

			simulator, err := NewFullDuplexSimulator(
				srv.url, &FullDuplexSimulatorOption{
					SendingConnectionMaxAliveTime:         time.Millisecond * 700,
					TargetSendingConnectionMaxAliveTime:   time.Millisecond * 700,
					ReceivingConnectionMaxAliveTime:       time.Millisecond * 10000,
					TargetReceivingConnectionMaxAliveTime: time.Millisecond * 10000,
					MaxSingleRequestSendingBytes:          1024 * 1024,
					MaxSingleRequestReceivingBytes:        1024 * 1024,
				},
			)
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
				sent, received, restarts := simulator.GetStats()
				t.Errorf(fmt.Sprintf("expected: %d, actual: %d, sent: %d, received: %d, restarts: %d", 3000*5, l, sent, received, restarts))
				server_recv_count := srv.recv_count
				server_send_count := srv.send_count
				t.Errorf(fmt.Sprintf("server recv count: %d, server send count: %d", server_recv_count, server_send_count))
				// to find which one is missing
				// for i := 0; i < 3000; i++ {
				// 	if !strings.Contains(recved.String(), fmt.Sprintf("%05d", i)) {
				// 		t.Errorf(fmt.Sprintf("missing: %d", i))
				// 	}
				// }
			}
		}()
	}

	wg.Wait()
}

func TestFullDuplexSimulator_MultipleTransactions(t *testing.T) {
	log.SetShowLog(false)
	defer log.SetShowLog(true)

	// avoid too many test cases, it will cause too many goroutines
	// finally, os will run into busy, and requests can not be handled correctly in time
	const NUM_CASES = 30

	w := sync.WaitGroup{}
	w.Add(NUM_CASES)

	for j := 0; j < NUM_CASES; j++ {
		// j := j
		go func() {
			defer w.Done()

			srv, err := server()
			if err != nil {
				t.Fatal(err)
			}
			defer srv.Stop()

			time.Sleep(time.Second)

			simulator, err := NewFullDuplexSimulator(
				srv.url, &FullDuplexSimulatorOption{
					SendingConnectionMaxAliveTime:         time.Millisecond * 700,
					TargetSendingConnectionMaxAliveTime:   time.Millisecond * 700,
					ReceivingConnectionMaxAliveTime:       time.Millisecond * 1000,
					TargetReceivingConnectionMaxAliveTime: time.Millisecond * 1000,
					MaxSingleRequestSendingBytes:          1024 * 1024,
					MaxSingleRequestReceivingBytes:        1024 * 1024,
				},
			)
			if err != nil {
				t.Fatal(err)
			}

			l := int32(0)

			dones := make(map[int]func())
			dones_lock := sync.Mutex{}

			buf := bytes.Buffer{}
			simulator.On(func(data []byte) {
				debugging.PossibleBlocking(
					func() any {
						atomic.AddInt32(&l, int32(len(data)))

						buf.Write(data)

						bytes := buf.Bytes()
						buf.Reset()

						i := 0
						for i < len(bytes) {
							num, err := strconv.Atoi(string(bytes[i : i+5]))
							if err != nil {
								t.Fatalf("invalid data: %s", string(bytes))
							}

							dones_lock.Lock()

							if done, ok := dones[num]; ok {
								done()
							} else {
								t.Fatalf("done not found: %d", num)
							}

							dones_lock.Unlock()

							i += 5
						}

						if buf.Len() != i {
							// write the rest of the data
							b := make([]byte, len(bytes)-i)
							copy(b, bytes[i:])
							buf.Write(b)
						}

						return nil
					},
					time.Second*1,
					func() {
						t.Fatal("possible blocking triggered")
					},
				)
			})

			wg := sync.WaitGroup{}
			wg.Add(100)

			for i := 0; i < 100; i++ {
				i := i
				time.Sleep(time.Millisecond * 20)
				go func() {
					done, err := simulator.StartTransaction()
					if err != nil {
						t.Fatal(err)
					}

					dones_lock.Lock()
					dones[i] = func() {
						done()
						wg.Done()
					}
					dones_lock.Unlock()

					if err := simulator.Send([]byte(fmt.Sprintf("%05d", i))); err != nil {
						t.Fatal(err)
					}
				}()
			}

			// time.AfterFunc(time.Second*5, func() {
			// 	// fmt.Println("server recv count: ", srv.recv_count, "server send count: ", srv.send_count, "j: ", j,
			// 	// 	"server recv request: ", srv.recv_request, "server send request: ", srv.send_request)
			// 	// fmt.Println("current_recv_request_id: ", srv.current_recv_request_id, "current_send_request_id: ", srv.current_send_request_id)
			// })

			wg.Wait()

			if l != 100*5 {
				sent, received, restarts := simulator.GetStats()
				t.Errorf(fmt.Sprintf("expected: %d, actual: %d, sent: %d, received: %d, restarts: %d", 100*5, l, sent, received, restarts))
			}
		}()
	}

	w.Wait()
}

func TestFullDuplexSimulator_SendLargeData(t *testing.T) {
	log.SetShowLog(false)
	defer log.SetShowLog(true)

	srv, err := server()
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Stop()

	time.Sleep(time.Second)

	l := 0

	simulator, err := NewFullDuplexSimulator(
		srv.url, &FullDuplexSimulatorOption{
			SendingConnectionMaxAliveTime:         time.Millisecond * 70000,
			TargetSendingConnectionMaxAliveTime:   time.Millisecond * 70000,
			ReceivingConnectionMaxAliveTime:       time.Millisecond * 100000,
			TargetReceivingConnectionMaxAliveTime: time.Millisecond * 100000,
			MaxSingleRequestSendingBytes:          5 * 1024 * 1024,
			MaxSingleRequestReceivingBytes:        5 * 1024 * 1024,
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	simulator.On(func(data []byte) {
		l += len(data)
	})

	done, err := simulator.StartTransaction()
	if err != nil {
		t.Fatal(err)
	}
	defer done()

	for i := 0; i < 300; i++ { // 300MB, this process should be done in 20 seconds
		if err := simulator.Send([]byte(strings.Repeat("a", 1024*1024))); err != nil {
			t.Fatal(err)
		}
	}

	time.Sleep(time.Second * 5)

	if l != 300*1024*1024 { // 300MB
		t.Fatal(fmt.Sprintf("expected: %d, actual: %d", 300*1024*1024, l))
	}
}
