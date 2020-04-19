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
func (c *OccurenceCounter) Values() map[uint32]int {
	c.mux.RLock()
	defer c.mux.RUnlock()

	// be able to return a current state of the mutex
	// in copying the values into a new map

	theMap := map[uint32]int{}
	for k, v := range c.v {
		theMap[k] = v
	}
	return theMap

}
func (c *OccurenceCounter) Clear() {
	c.mux.Lock()
	c.v = make(map[uint32]int)
	c.mux.Unlock()
}

func CountOccurance(i uint32, m *OccurenceCounter) {
	m.Inc(i)
}

func main() {
	ticker := time.NewTicker(5 * time.Second)
	quitTicker := make(chan struct{})
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	oc := OccurenceCounter{v: make(map[uint32]int)}
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	stopMainLoop := false

	// have an interval to write to postgres
	go func(mx *OccurenceCounter) {
		for {
			select {
			case <-ticker.C:
				// main logic
				val := mx.Values()
				total := 0
				for _, v := range val {
					total = total + v
				}


				fmt.Println("total:",total)
				// cleaning up the mutex
				mx.Clear()
			case <-quitTicker:
				stopMainLoop = true
				fmt.Println("quitting...")
				fmt.Println("If there is still data in OccurenceCounter write it away")
				val := mx.Values()
				fmt.Println(val)
				ticker.Stop()
				return
			}

		}
	}(&oc)

	// main loop
	for {

		if stopMainLoop == true{
			break
		}
		theInt := uint32(rand.Intn(100))
		go CountOccurance(theInt, &oc)
		go func() {
			sig := <-sigs
			fmt.Println()
			fmt.Println(sig)
			done <- true
		}()

		// where to exit the process so that no data is being lost?

		go func(){
			<-done
			fmt.Println("exiting")
			time.Sleep(5 * time.Second)
			close(quitTicker)

		}()



	}
}
