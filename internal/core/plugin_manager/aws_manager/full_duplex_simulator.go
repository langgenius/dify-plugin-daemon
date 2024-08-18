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

	// how many transactions are alive
	alive_transactions int32

	// total transactions
	total_transactions int32

	// sending_connection_timeline_lock
	sending_connection_timeline_lock sync.Mutex
	// sending pipeline
	sending_pipeline *io.PipeWriter

	// receiving_connection_timeline_lock
	receiving_connection_timeline_lock sync.Mutex
	// receiving reader
	receiving_reader io.ReadCloser

	// max retries
	max_retries int

	// is sending connection alive
	sending_connection_alive int32

	// is receiving connection alive
	receiving_connection_alive int32

	// listener for data
	listeners []func(data []byte)

	// mutex for listeners
	listeners_mu sync.RWMutex

	// request id
	request_id string

	// http client
	client *http.Client
}

func NewFullDuplexSimulator(baseurl string) (*FullDuplexSimulator, error) {
	u, err := url.Parse(baseurl)
	if err != nil {
		return nil, err
	}

	return &FullDuplexSimulator{
		baseurl:     u,
		max_retries: 10,
		request_id:  strings.RandomString(32),

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
func (s *FullDuplexSimulator) Send(data []byte) error {
	if atomic.LoadInt32(&s.sending_connection_alive) == 0 {
		return errors.New("sending connection is not alive")
	}

	writer := s.sending_pipeline
	if writer == nil {
		return errors.New("sending pipeline is not alive")
	}

	if _, err := writer.Write(data); err != nil {
		return err
	}

	return nil
}

func (s *FullDuplexSimulator) On(f func(data []byte)) {
	s.listeners_mu.Lock()
	defer s.listeners_mu.Unlock()
	s.listeners = append(s.listeners, f)
}

// start a transaction
// returns a function to stop the transaction
func (s *FullDuplexSimulator) StartTransaction() (func(), error) {
	// start a transaction
	atomic.AddInt32(&s.alive_transactions, 1)
	atomic.AddInt32(&s.total_transactions, 1)

	// start sending connection
	if err := s.startSendingConnection(); err != nil {
		return nil, err
	}

	// start receiving connection
	if err := s.startReceivingConnection(); err != nil {
		return nil, err
	}

	return s.stopTransaction, nil
}

func (s *FullDuplexSimulator) stopTransaction() {
	// close if no transaction is alive
	if atomic.AddInt32(&s.alive_transactions, -1) == 0 {
		s.stopSendingConnection()
		s.stopReceivingConnection()
	}
}

func (s *FullDuplexSimulator) startSendingConnection() error {
	if atomic.LoadInt32(&s.sending_connection_alive) == 1 {
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

	pr, pw := io.Pipe()

	req, err := http.NewRequest("POST", u, pr)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "octet-stream")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("x-dify-plugin-request-id", s.request_id)

	routine.Submit(func() {
		s.sendingConnectionRoutine(req)
	})

	// mark sending connection as alive
	atomic.StoreInt32(&s.sending_connection_alive, 1)

	// set the sending pipeline
	s.sending_pipeline = pw

	return nil
}

func (s *FullDuplexSimulator) sendingConnectionRoutine(origin_req *http.Request) {
	failed_times := 0
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		time.AfterFunc(5*time.Second, func() {
			cancel()
		})
		req := origin_req.Clone(ctx)
		req = req.WithContext(ctx)
		resp, err := s.client.Do(req)

		if err != nil {
			failed_times++
			if failed_times > s.max_retries {
				log.Error("failed to establish sending connection: %v", err)
				s.stopSendingConnection()
				return
			}

			log.Error("failed to establish sending connection: %v", err)
			continue
		}

		defer resp.Body.Close()

		// mark sending connection as dead
		atomic.StoreInt32(&s.sending_connection_alive, 0)

		s.sending_connection_timeline_lock.Lock()
		defer s.sending_connection_timeline_lock.Unlock()

		// close the sending pipeline
		if s.sending_pipeline != nil {
			s.sending_pipeline.Close()
			s.sending_pipeline = nil
		}
	}
}

func (s *FullDuplexSimulator) stopSendingConnection() error {
	if atomic.LoadInt32(&s.sending_connection_alive) == 0 {
		return nil
	}

	s.sending_connection_timeline_lock.Lock()
	defer s.sending_connection_timeline_lock.Unlock()

	// close the sending pipeline
	if s.sending_pipeline != nil {
		s.sending_pipeline.Close()
		s.sending_pipeline = nil
	}

	// mark sending connection as dead
	atomic.StoreInt32(&s.sending_connection_alive, 0)

	return nil
}

func (s *FullDuplexSimulator) startReceivingConnection() error {
	if atomic.LoadInt32(&s.receiving_connection_alive) == 1 {
		return nil
	}

	// lock the receiving connection
	s.receiving_connection_timeline_lock.Lock()
	defer s.receiving_connection_timeline_lock.Unlock()

	// start a new receiving connection
	u, err := url.JoinPath(s.baseurl.String(), "/response")
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "octet-stream")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("x-dify-plugin-request-id", s.request_id)

	req = req.Clone(context.Background())
	resp, err := s.client.Do(req)
	if err != nil {
		return errors.Join(err, errors.New("failed to establish receiving connection"))
	}

	routine.Submit(func() {
		s.receivingConnectionRoutine(req, resp.Body)
	})

	// mark receiving connection as alive
	atomic.StoreInt32(&s.receiving_connection_alive, 1)

	return nil
}

func (s *FullDuplexSimulator) receivingConnectionRoutine(req *http.Request, reader io.ReadCloser) {
	failed_times := 0
	for {
		s.receiving_reader = reader
		recved_pong := false
		buf := make([]byte, 0)
		buf_len := 0

		for {
			data := make([]byte, 1024)
			n, err := reader.Read(data)
			if err != nil {
				break
			}

			// check if pong\n is at the beginning of the data
			if !recved_pong {
				data = append(data, buf[:buf_len]...)
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

			for _, listener := range s.listeners[:] {
				listener(data[:n])
			}
		}

		s.receiving_reader = nil

		s.receiving_connection_timeline_lock.Lock()
		if atomic.LoadInt32(&s.receiving_connection_alive) == 0 {
			s.receiving_connection_timeline_lock.Unlock()
			return
		}
		s.receiving_connection_timeline_lock.Unlock()

		req = req.Clone(context.Background())
		resp, err := s.client.Do(req)
		if err != nil {
			failed_times++
			if failed_times > s.max_retries {
				log.Error("failed to establish receiving connection: %v", err)
				s.stopReceivingConnection()
				return
			}

			log.Error("failed to establish receiving connection: %v", err)
			continue
		}

		reader = resp.Body
	}
}

func (s *FullDuplexSimulator) stopReceivingConnection() {
	if atomic.LoadInt32(&s.receiving_connection_alive) == 0 {
		return
	}

	// mark receiving connection as dead
	atomic.StoreInt32(&s.receiving_connection_alive, 0)

	s.receiving_connection_timeline_lock.Lock()
	defer s.receiving_connection_timeline_lock.Unlock()

	// close the receiving reader
	reader := s.receiving_reader
	if reader != nil {
		reader.Close()
	}
}
