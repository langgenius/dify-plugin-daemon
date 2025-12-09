package routine

import (
	"context"
	"runtime"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/getsentry/sentry-go"
	routinelabels "github.com/langgenius/dify-plugin-daemon/pkg/routine"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/log"
	"github.com/panjf2000/ants/v2"
)

var (
	p *ants.Pool
	l sync.Mutex
)

func IsInit() bool {
	l.Lock()
	defer l.Unlock()
	return p != nil
}

func InitPool(size int, sentryOption ...sentry.ClientOptions) {
	l.Lock()
	defer l.Unlock()
	if p != nil {
		return
	}
	log.Info("init routine pool, size: %d", size)
	p, _ = ants.NewPool(size, ants.WithNonblocking(false))

	if len(sentryOption) > 0 {
		if err := sentry.Init(sentryOption[0]); err != nil {
			log.Error("init sentry failed, error: %v", err)
		}
	}
}

func Submit(labels routinelabels.Labels, f func()) {
	if labels == nil {
		labels = routinelabels.Labels{}
	}

	p.Submit(func() {
		label := []string{
			"LaunchedAt", time.Now().Format(time.RFC3339),
		}
		if len(labels) > 0 {
			for k, v := range labels {
				label = append(label, string(k), v)
			}
		}
		pprof.Do(context.Background(), pprof.Labels(label...), func(ctx context.Context) {
			defer sentry.Recover()
			defer func() {
				if err := recover(); err != nil {
					log.Error("panic in routine: %v", err)
					// get stack trace
					buf := make([]byte, 1024*1024)
					n := runtime.Stack(buf, false)
					log.Error("stack trace %v: %s", err, string(buf[:n]))
				}
			}()
			f()
		})
	})
}

func WithMaxRoutine(maxRoutine int, tasks []func(), onFinish ...func()) {
	if maxRoutine <= 0 {
		maxRoutine = 1
	}

	if maxRoutine > len(tasks) {
		maxRoutine = len(tasks)
	}

	Submit(routinelabels.Labels{
		routinelabels.RoutineLabelKeyModule: "routine",
		routinelabels.RoutineLabelKeyMethod: "WithMaxRoutine",
	}, func() {
		wg := sync.WaitGroup{}
		taskIndex := int32(0)

		for i := 0; i < maxRoutine; i++ {
			wg.Add(1)
			Submit(nil, func() {
				defer wg.Done()
				currentIndex := atomic.AddInt32(&taskIndex, 1)
				for currentIndex <= int32(len(tasks)) {
					task := tasks[currentIndex-1]
					task()
					currentIndex = atomic.AddInt32(&taskIndex, 1)
				}
			})
		}

		wg.Wait()

		if len(onFinish) > 0 {
			onFinish[0]()
		}
	})
}

type PoolStatus struct {
	Free  int `json:"free"`
	Busy  int `json:"busy"`
	Total int `json:"total"`
}

func FetchRoutineStatus() *PoolStatus {
	return &PoolStatus{
		Free:  p.Free(),
		Busy:  p.Running(),
		Total: p.Cap(),
	}
}
