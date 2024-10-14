package routine

import (
	"sync"
	"sync/atomic"

	"github.com/langgenius/dify-plugin-daemon/internal/utils/log"
	"github.com/panjf2000/ants"
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

func InitPool(size int) {
	l.Lock()
	defer l.Unlock()
	if p != nil {
		return
	}
	log.Info("init routine pool, size: %d", size)
	p, _ = ants.NewPool(size, ants.WithNonblocking(false))
}

func Submit(f func()) {
	p.Submit(f)
}

func WithMaxRoutine(max_routine int, tasks []func(), on_finish ...func()) {
	if max_routine <= 0 {
		max_routine = 1
	}

	if max_routine > len(tasks) {
		max_routine = len(tasks)
	}

	Submit(func() {
		wg := sync.WaitGroup{}
		task_index := int32(0)

		for i := 0; i < max_routine; i++ {
			wg.Add(1)
			Submit(func() {
				defer wg.Done()
				current_index := atomic.AddInt32(&task_index, 1)

				if current_index >= int32(len(tasks)) {
					return
				}

				for current_index < int32(len(tasks)) {
					task := tasks[current_index]
					task()
					current_index++
				}
			})
		}

		wg.Wait()

		if len(on_finish) > 0 {
			on_finish[0]()
		}
	})
}
