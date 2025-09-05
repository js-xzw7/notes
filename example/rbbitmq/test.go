package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	count := 10
	sum := 100
	wg := sync.WaitGroup{}

	c := int64(0)
	ch := make(chan struct{}, count)

	for i := 0; i < sum; i++ {
		wg.Add(1)
		ch <- struct{}{}

		go func(j int) {
			defer wg.Done()
			<-ch
			fmt.Printf("已完成第[%d]个携程\n", j)
			atomic.AddInt64(&c, 1)
			time.Sleep(time.Second * 1)
		}(i)
	}

	wg.Wait()
	fmt.Println("统计完成数量", c)
}
