package util

import (
	"log"
	"time"
)

func GetQPS(f func(n int), n int) {
	start := time.Now()
	f(n)
	end := time.Now()
	seconds := end.Sub(start).Seconds()
	log.Printf("Qps: %f\n", float64(n)/seconds)
}

func GetQPSAsnyc(f func(i int), runTimes int, concurrency int) {
	c := make(chan int, concurrency)
	start := time.Now()
	for i := 0; i < runTimes; i++ {
		c <- i
		go func(i int) {
			defer func() {
				<-c
			}()
			f(i)
		}(i)
	}
	end := time.Now()
	seconds := end.Sub(start).Seconds()
	log.Printf("Qps: %f\n", float64(runTimes)/seconds)
}
