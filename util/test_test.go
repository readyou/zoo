package util

import (
	"context"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/semaphore"
	"sync"
	"testing"
	"time"
)

func TestBaseQPS(t *testing.T) {
	qps := GetQPS(func(n int) {
		for i := 0; i < n; i++ {
			// sleep不精确
			time.Sleep(time.Millisecond * 10)
		}
	}, 1000)
	assert.True(t, qps > 80 && qps < 100)
}

func TestSemQPS(t *testing.T) {
	sem := semaphore.NewWeighted(1000)
	n := 10000
	qps := GetQPS(func(n int) {
		wg := sync.WaitGroup{}
		wg.Add(n)
		for i := 0; i < n; i++ {
			go func() {
				sem.Acquire(context.Background(), 1)
				defer sem.Release(1)
				time.Sleep(time.Millisecond * 10)
				wg.Done()
			}()
		}
		wg.Wait()
	}, n)
	assert.True(t, qps > 80000 && qps < 100000)
}

func TestSemQPS2(t *testing.T) {
	// 防止太多的Go程，不然也会耗费额外的资源导致性能下降
	maxGo := semaphore.NewWeighted(10000)
	sem := semaphore.NewWeighted(10)
	n := 1000000
	qps := GetQPS(func(n int) {
		wg := sync.WaitGroup{}
		wg.Add(n)
		for i := 0; i < n; i++ {
			maxGo.Acquire(context.Background(), 1)
			go func() {
				defer maxGo.Release(1)
				sem.Acquire(context.Background(), 1)
				defer sem.Release(1)
				wg.Done()
			}()
		}
		wg.Wait()
	}, n)
	assert.True(t, qps > 250000)
}

func TestGetQPSAsnyc(t *testing.T) {
	n := 10000
	qps := GetQPSAsnyc(func(i int) {
		time.Sleep(time.Millisecond * 10)
	}, n, 1000)
	assert.True(t, qps > 80000 && qps < 100000)

	qps = GetQPSAsnyc(func(i int) {
	}, n, 1000)
	// 用chan实现的并发控制比semaphore实现的，性能要好一些（高并发情况下）
	// qps在10万以下的话，性能差不多，semaphore代码更优雅一点
	assert.True(t, qps > 1000000)
}
