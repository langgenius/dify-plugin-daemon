package main

import (
	"fmt"
	"time"
)

func main() {
	ch := c()
	for i := range ch {
		if i == 20 {
			break
		}
		fmt.Println(i)
	}

	for {
		time.Sleep(1 * time.Second)
	}
}

func c() <-chan int {
	c := make(chan int)
	go func() {
		for i := 0; i < 10000; i++ {
			fmt.Println("send", i)
			c <- i
		}

		close(c)
	}()
	return c
}
