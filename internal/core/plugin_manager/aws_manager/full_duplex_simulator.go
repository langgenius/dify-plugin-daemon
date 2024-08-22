package aws_manager

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/routine"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/strings"
)

// Full duplex simulator, using http protocol to simulate the full duplex communication
// 1. during free time, no connection will be established
// 2. when there is a virtual connection need to be established, 2 http transactions will be sent to the server
// 3. one is used to send data chunk by chunk to simulate the data stream and the other is used to receive data using event stream
// 4. after all data is sent, the connection will be closed to reduce the network traffic
//
// When http connection is closed, the simulator will restart it immediately until it has reached max_retries
type FullDuplexSimulator struct {
	baseurl *url.URL

	// single connection max alive time
	sending_connection_max_alive_time   time.Duration
	receiving_connection_max_alive_time time.Duration

	// how many transactions are alive
	alive_transactions int32

	// total transactions
	total_transactions int32

	// connection restarts
	connection_restarts int32

	// sent bytes
	sent_bytes int64
	// received bytes
	received_bytes int64

	// sending_connection_timeline_lock
	sending_connection_timeline_lock sync.Mutex
	// sending pipeline
	sending_pipeline *io.PipeWriter
	// sending pipe lock
	sending_pipe_lock sync.RWMutex

	// receiving_connection_timeline_lock
	receiving_connection_timeline_lock sync.Mutex
	// receiving context
	receiving_cancel context.CancelFunc
	// receiving context lock
	receiving_cancel_lock sync.Mutex

	// max retries
	max_retries int

	// request id
	request_id string

	// latest routine id
	latest_routine_id string

	// is sending connection alive
	sending_connection_alive         int32
	sending_routine_lock             sync.Mutex
	virtual_sending_connection_alive int32

	// receiving routine lock
	receiving_routine_lock sync.Mutex
	// is receiving connection alive
	virtual_receiving_connection_alive int32

	// listener for data
	listeners []func(data []byte)

	// mutex for listeners
	listeners_lock sync.RWMutex

	// http client
	client *http.Client
}

func NewFullDuplexSimulator(
	baseurl string,
	sending_connection_max_alive_time time.Duration,
	receiving_connection_max_alive_time time.Duration,
) (*FullDuplexSimulator, error) {
	u, err := url.Parse(baseurl)
	if err != nil {
		return nil, err
	}

	return &FullDuplexSimulator{
		baseurl:                             u,
		sending_connection_max_alive_time:   sending_connection_max_alive_time,
		receiving_connection_max_alive_time: receiving_connection_max_alive_time,
		max_retries:                         10,
		request_id:                          strings.RandomString(32),

		// using keep alive to reduce the connection reset
		client: &http.Client{
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 15 * time.Second,
				}).Dial,
				IdleConnTimeout: 120 * time.Second,
			},
		},
	}, nil
}

// send data to server
func (s *FullDuplexSimulator) Send(data []byte, timeout ...time.Duration) error {
	timeout_duration := time.Second * 10
	if len(timeout) > 0 {
		timeout_duration = timeout[0]
	}

	started := time.Now()

	for time.Since(started) < timeout_duration {
		if atomic.LoadInt32(&s.sending_connection_alive) != 1 {
			time.Sleep(time.Millisecond * 50)
			continue
		}

		s.sending_pipe_lock.Lock()
		writer := s.sending_pipeline
		if writer == nil {
			time.Sleep(time.Millisecond * 50)
			s.sending_pipe_lock.Unlock()
			continue
		}

		if n, err := writer.Write(data); err != nil {
			time.Sleep(time.Millisecond * 50)
			s.sending_pipe_lock.Unlock()
			continue
		} else {
			atomic.AddInt64(&s.sent_bytes, int64(n))
		}

		s.sending_pipe_lock.Unlock()
		return nil
	}

	return errors.New("send data timeout")
}

func (s *FullDuplexSimulator) On(f func(data []byte)) {
	s.listeners_lock.Lock()
	defer s.listeners_lock.Unlock()
	s.listeners = append(s.listeners, f)
}

// start a transaction
// returns a function to stop the transaction
func (s *FullDuplexSimulator) StartTransaction() (func(), error) {
	// start a transaction
	if atomic.AddInt32(&s.alive_transactions, 1) == 1 {
		// increase connection restarts
		atomic.AddInt32(&s.connection_restarts, 1)

		// reset request id
		routine_id := strings.RandomString(32)

		// update latest request id
		s.latest_routine_id = routine_id

		// start sending connection
		if err := s.startSendingConnection(routine_id); err != nil {
			return nil, err
		}

		// start receiving connection
		if err := s.startReceivingConnection(routine_id); err != nil {
			s.stopSendingConnection()
			return nil, err
		}
	}

	atomic.AddInt32(&s.total_transactions, 1)

	return s.stopTransaction, nil
}

func (s *FullDuplexSimulator) stopTransaction() {
	// close if no transaction is alive
	if atomic.AddInt32(&s.alive_transactions, -1) == 0 {
		s.stopSendingConnection()
		s.stopReceivingConnection()
	}
}

func (s *FullDuplexSimulator) startSendingConnection(routine_id string) error {
	// if virtual sending connection is already alive, do nothing
	if !atomic.CompareAndSwapInt32(&s.virtual_sending_connection_alive, 0, 1) {
		return nil
	}

	// lock the sending connection
	s.sending_connection_timeline_lock.Lock()
	defer s.sending_connection_timeline_lock.Unlock()

	// start a new sending connection
	u, err := url.JoinPath(s.baseurl.String(), "/invoke")
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "octet-stream")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("x-dify-plugin-request-id", s.request_id)

	routine.Submit(func() {
		s.sendingConnectionRoutine(req, routine_id)
	})

	return nil
}

func (s *FullDuplexSimulator) sendingConnectionRoutine(origin_req *http.Request, routine_id string) {
	// lock the sending routine, to avoid there are multiple routines trying to establish the sending connection
	s.sending_routine_lock.Lock()

	// cancel the sending routine
	defer s.sending_routine_lock.Unlock()

	failed_times := 0
	for atomic.LoadInt32(&s.virtual_sending_connection_alive) == 1 {
		// check if the request id is the latest one, avoid this routine being used by a old request
		if routine_id != s.latest_routine_id {
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		time.AfterFunc(s.sending_connection_max_alive_time, func() {
			// reached max alive time, remove pipe writer
			s.sending_pipe_lock.Lock()
			if s.sending_pipeline != nil {
				s.sending_pipeline.Close()
				s.sending_pipeline = nil
			}
			s.sending_pipe_lock.Unlock()
			time.AfterFunc(time.Second, cancel)
		})

		req := origin_req.Clone(ctx)
		pr, pw := io.Pipe()
		s.sending_pipe_lock.Lock()
		req.Body = pr
		s.sending_pipeline = pw
		s.sending_pipe_lock.Unlock()
		req = req.WithContext(ctx)

		// mark sending connection as alive
		atomic.StoreInt32(&s.sending_connection_alive, 1)

		resp, err := s.client.Do(req)
		if err != nil {
			atomic.StoreInt32(&s.sending_connection_alive, 0)

			// if virtual sending connection is not alive, clear the sending pipeline and return
			if atomic.LoadInt32(&s.virtual_sending_connection_alive) == 0 {
				// clear the sending pipeline
				s.sending_pipe_lock.Lock()
				if s.sending_pipeline != nil {
					s.sending_pipeline.Close()
					s.sending_pipeline = nil
				}
				s.sending_pipe_lock.Unlock()
				return
			}

			failed_times++
			if failed_times > s.max_retries {
				log.Error("failed to establish sending connection: %v", err)
				s.stopSendingConnection()
				return
			}

			log.Error("failed to establish sending connection: %v", err)
		} else {
			defer resp.Body.Close()
		}

		// mark sending connection as dead
		atomic.StoreInt32(&s.sending_connection_alive, 0)

		s.sending_pipe_lock.Lock()
		// close the sending pipeline
		if s.sending_pipeline != nil {
			s.sending_pipeline.Close()
			s.sending_pipeline = nil
		}
		s.sending_pipe_lock.Unlock()
	}
}

func (s *FullDuplexSimulator) stopSendingConnection() error {
	if !atomic.CompareAndSwapInt32(&s.virtual_sending_connection_alive, 1, 0) {
		return nil
	}

	s.sending_connection_timeline_lock.Lock()
	defer s.sending_connection_timeline_lock.Unlock()

	s.sending_pipe_lock.Lock()
	defer s.sending_pipe_lock.Unlock()

	// close the sending pipeline
	if s.sending_pipeline != nil {
		s.sending_pipeline.Close()
		s.sending_pipeline = nil
	}

	// mark sending connection as dead
	atomic.StoreInt32(&s.virtual_sending_connection_alive, 0)

	return nil
}

func (s *FullDuplexSimulator) startReceivingConnection(request_id string) error {
	// if virtual receiving connection is already alive, do nothing
	if !atomic.CompareAndSwapInt32(&s.virtual_receiving_connection_alive, 0, 1) {
		return nil
	}

	// lock the receiving connection
	s.receiving_connection_timeline_lock.Lock()
	defer s.receiving_connection_timeline_lock.Unlock()

	routine.Submit(func() {
		s.receivingConnectionRoutine(request_id)
	})

	return nil
}

func (s *FullDuplexSimulator) receivingConnectionRoutine(routine_id string) {
	// lock the receiving routine, to avoid there are multiple routines trying to establish the receiving connection
	s.receiving_routine_lock.Lock()
	// cancel the receiving routine
	defer s.receiving_routine_lock.Unlock()

	for atomic.LoadInt32(&s.virtual_receiving_connection_alive) == 1 {
		// check if the request id is the latest one, avoid this routine being used by a old request
		if routine_id != s.latest_routine_id {
			return
		}

		recved_pong := false
		buf := make([]byte, 0)
		buf_len := 0

		// start a new receiving connection
		u, err := url.JoinPath(s.baseurl.String(), "/response")
		if err != nil {
			continue
		}

		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			continue
		}
		req.Header.Set("Content-Type", "octet-stream")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("x-dify-plugin-request-id", s.request_id)

		ctx, cancel := context.WithCancel(context.Background())
		req = req.Clone(ctx)
		resp, err := s.client.Do(req)
		if err != nil {
			continue
		}

		s.receiving_cancel_lock.Lock()
		s.receiving_cancel = cancel
		s.receiving_cancel_lock.Unlock()

		time.AfterFunc(s.receiving_connection_max_alive_time, func() {
			cancel()
			resp.Body.Close()
		})

		reader := resp.Body
		for {
			data := make([]byte, 1024)
			n, err := reader.Read(data)
			if n != 0 {
				// check if pong\n is at the beginning of the data
				if !recved_pong {
					data = append(buf[:buf_len], data[:n]...)
					buf = make([]byte, 0)
					buf_len = 0

					if n >= 5 {
						if string(data[:5]) == "pong\n" {
							recved_pong = true
							// remove pong\n from the beginning of the data
							data = data[5:]
							n -= 5
						} else {
							// not pong\n, break
							break
						}
					} else if n < 5 {
						// save the data to the buffer
						buf = append(buf, data[:n]...)
						buf_len += n
						continue
					}
				}
			}

			for _, listener := range s.listeners[:] {
				listener(data[:n])
			}

			atomic.AddInt64(&s.received_bytes, int64(n))

			if err != nil {
				break
			}
		}
	}
}

func (s *FullDuplexSimulator) stopReceivingConnection() {
	if !atomic.CompareAndSwapInt32(&s.virtual_receiving_connection_alive, 1, 0) {
		return
	}

	// cancel the receiving context
	s.receiving_cancel_lock.Lock()
	if s.receiving_cancel != nil {
		s.receiving_cancel()
	}
	s.receiving_cancel_lock.Unlock()
}

// GetStats, returns the sent and received bytes
func (s *FullDuplexSimulator) GetStats() (sent_bytes, received_bytes int64, connection_restarts int32) {
	return atomic.LoadInt64(&s.sent_bytes), atomic.LoadInt64(&s.received_bytes), atomic.LoadInt32(&s.connection_restarts)
}
