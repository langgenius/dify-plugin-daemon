package main

import (
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/langgenius/dify-plugin-daemon/internal/types/entities"
)

func main() {
	response := entities.NewInvocationResponse[string](1024)

	random_string := func() string {
		return fmt.Sprintf("%d", rand.Intn(100000))
	}

	traffic := new(int64)

	go func() {
		for {
			response.Write(random_string())
		}
	}()

	go func() {
		for {
			response.Write(random_string())
		}
	}()

	go func() {
		for response.Next() {
			atomic.AddInt64(traffic, 1)
			_, err := response.Read()
			if err != nil {
				fmt.Println(err)
				break
			}
		}
	}()

	go func() {
		for range time.NewTicker(time.Second).C {
			fmt.Printf("Traffic: %d, Unsolved: %d\n", atomic.LoadInt64(traffic), response.Size())
			atomic.StoreInt64(traffic, 0)
		}
	}()

	select {}
}
