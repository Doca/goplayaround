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

const(
	maxPoolSize = 1000
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


	// be able to return a current state of the mutex
	// in copying the values into a new map

	total := 0
	for _, v := range c.v {
		total = total + v
	}

	// Simulates the save into PG
	fmt.Println("total:",total)

	// cleaning up the mutex
	c.clear()
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
	//done := make(chan bool, 1)
	oc := OccurenceCounter{v: make(map[uint32]int)}
	stopMainLoop := false


	wg := &sync.WaitGroup{}

	// have an interval to write to postgres
	wg.Add(1)
	go func(mx *OccurenceCounter) {
		defer wg.Done()
		//defer func() {
		//	fmt.Println(" initial goroutine is complete")
		//	wg.Done()
		//}()

		for {
			select {
			case <-ticker.C:
				// main logic
				mx.Save()
			case <-quitTicker:
				stopMainLoop = true
				fmt.Println("quitting...")
				fmt.Println("If there is still data in OccurenceCounter write it away")
				//val := mx.Values()
				//fmt.Println(val)
				ticker.Stop()


				return
			}

		}
	}(&oc)

	// main loop
	//wp := workerpool.New(maxPoolSize)
	sem := make(chan int, maxPoolSize)
	//wp.Submit(func() {
	//	fmt.Println("Test123")
	//})

	fmt.Println("Starting process...")



	go func(){
		for {
			wg.Add(1)
			sem <-1
			go func(){
				defer wg.Done()
				theInt := uint32(rand.Intn(100))
				CountOccurance(theInt, &oc)
				<-sem
			}()
		}
	}()

	<- sigs
	fmt.Println("wait for waitgroup")
	close(quitTicker)

	wg.Wait()



	// One final flush of anything in oc.
	oc.Save()

	fmt.Println("Shutdown success...")

	//for {
	//
	//	if stopMainLoop == true {
	//		break
	//	}
	//
	//	theInt := uint32(rand.Intn(100))
	//	// Fine
	//
	//	wp.Submit(func(){CountOccurance(theInt, &oc)})
	//	//wp.Submit(func(){
	//	//	sig := <-sigs
	//	//	fmt.Println()
	//	//	fmt.Println(sig)
	//	//	done <- true
	//	//})
	//
	//	// where to exit the process so that no data is being lost?
	//
	//	//wp.Submit(func(){
	//	//	<-done
	//	//	fmt.Println("exiting")
	//	//	time.Sleep(5 * time.Second)
	//	//	close(quitTicker)
	//	//
	//	//})
	//
	//
	//
	//}
}
