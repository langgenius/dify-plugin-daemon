package lock

import (
	"fmt"
	"sync"
	"testing"
)

func TestHighGranularityLock(t *testing.T) {
	l := NewGranularityLock()

	data := []int{}
	add := func(key int) {
		l.Lock(fmt.Sprintf("%d", key))
		data[key]++
		l.Unlock(fmt.Sprintf("%d", key))
	}

	for i := 0; i < 1000; i++ {
		data = append(data, 0)
	}

	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			for j := 0; j < 1000; j++ {
				add(j)
			}
			wg.Done()
		}()
	}
	wg.Wait()

	for _, v := range data {
		if v != 1000 {
			t.Fatal("data not equal")
		}
	}
}
