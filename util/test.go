package util

import (
	"log"
	"sync"
	"time"
)

func GetQPS(f func(n int), n int) int {
	start := time.Now()
	f(n)
	end := time.Now()
	seconds := end.Sub(start).Seconds()
	qps := int(float64(n) / seconds)
	log.Printf("Qps: %d\n", qps)
	return qps
}

func GetQPSAsnyc(f func(i int), runTimes int, concurrency int) int {
	c := make(chan int, concurrency)
	start := time.Now()
	wg := sync.WaitGroup{}
	wg.Add(runTimes)
	for i := 0; i < runTimes; i++ {
		c <- i
		go func(i int) {
			defer func() {
				wg.Done()
				<-c
			}()
			f(i)
		}(i)
	}
	wg.Wait()
	end := time.Now()
	seconds := end.Sub(start).Seconds()
	qps := int(float64(runTimes) / seconds)
	log.Printf("Qps: %d\n", qps)
	return qps
}
