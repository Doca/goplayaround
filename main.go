package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type OccurenceCounter struct {
	v   map[uint32]int
	mux sync.Mutex
}

func (c *OccurenceCounter) Inc(key uint32) {
	c.mux.Lock()
	c.v[key]++
	c.mux.Unlock()
}

func (c *OccurenceCounter) Value(key uint32) int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.v[key]
}
func (c *OccurenceCounter) Values(key uint32) int {
	c.mux.Lock()
	defer c.mux.Unlock()
	return c.v[key]
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
	quit := make(chan struct{})
	oc := OccurenceCounter{v: make(map[uint32]int)}

	go func(mx *OccurenceCounter) {
		for {
			select {
			case <-ticker.C:
				fmt.Println(mx.Value(uint32(40)))
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}(&oc)
	for {
		theInt := uint32(rand.Intn(100))
		go CountOccurance(theInt, &oc)
	}
}
