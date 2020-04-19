package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// This won't work
// ERROR: race: limit on 8128 simultaneously alive goroutines is exceeded, dying
// exit status 66
// go run --race main.go

const (
	maxPoolSize = 20
)

type OccurenceCounter struct {
	v   map[uint32]int
	mux sync.RWMutex
}

func (c *OccurenceCounter) Inc(key uint32) {
	c.mux.Lock()
	c.v[key]++
	c.mux.Unlock()
}

func (c *OccurenceCounter) Value(key uint32) int {
	c.mux.RLock()
	defer c.mux.RUnlock()
	return c.v[key]
}

func (c *OccurenceCounter) Save() {
	c.mux.Lock()
	defer c.mux.Unlock()

	// 1. In Go, Locks are not re-entrant...

	total := c.agg()

	// Simulates the save into PG
	fmt.Println("Save into Postgres total:", total)

	// cleaning up the mutex
	c.clear()
}

func (c *OccurenceCounter) Total() int {
	c.mux.RLock()
	defer c.mux.RUnlock()

	return c.agg()
}

func (c *OccurenceCounter) agg() int {
	total := 0
	for _, v := range c.v {
		total = total + v
	}
	return total
}

func (c *OccurenceCounter) Clear() {
	c.mux.Lock()
	c.clear()
	c.mux.Unlock()
}

func (c *OccurenceCounter) clear() {
	c.v = make(map[uint32]int)
}

func CountOccurance(i uint32, m *OccurenceCounter) {
	m.Inc(i)
}

func main() {

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(2 * time.Second)
	quitTicker := make(chan struct{})
	flushDone := make(chan int)

	oc := &OccurenceCounter{v: make(map[uint32]int)}

	// Periodically write to Postgres.
	go func(mx *OccurenceCounter) {
		for {
			select {
			case <-ticker.C:
				// main logic
				mx.Save()
			case <-quitTicker:
				ticker.Stop()

				// Attempt one final save.
				mx.Save()

				// Signal final flush is completed.
				close(flushDone)
				return
			}

		}
	}(oc)

	// Bound worker-pool with max goroutines.
	sem := make(chan int, maxPoolSize)

	go func() {
		var myWG sync.WaitGroup
		for {
			select {
			// Upon sigterm, start teardown.
			case <-sigs:
				// Wait for all worker items to finish.
				myWG.Wait()

				// Ask for tear down ticker.
				close(quitTicker)
				return
			default:
				// Otherwise, continue to enqueue.
				myWG.Add(1)
				sem <- 1
				go func() {
					defer myWG.Done()
					enqueueTask(&myWG, oc)
					<-sem
				}()
			}
		}
	}()

	// Wait for final flush from ticker tear down.
	<-flushDone

	// Ensure zero result.
	fmt.Println("Final total: ", oc.Total())

	fmt.Println("Shutdown success...")
}

func enqueueTask(wg *sync.WaitGroup, oc *OccurenceCounter) {
	theInt := uint32(rand.Intn(100))
	CountOccurance(theInt, oc)
}
