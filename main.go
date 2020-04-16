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
func (c *OccurenceCounter) Clear(){
	c.mux.Lock()
	c.v = make(map[uint32]int)
	c.mux.Unlock()
}


func CountOccurance(i uint32, m *OccurenceCounter){
		m.Inc(i)
}


func main(){
	oc := OccurenceCounter{v: make(map[uint32]int)}
	tick := 0
	start := time.Now()
	for {
		theInt := uint32(rand.Intn(100))
		go CountOccurance(theInt,&oc)
		tick ++
		if tick == 300000{
			elapsed := time.Since(start)
			fmt.Println(elapsed)
			cleanStart := time.Now()
			oc.Clear()
			cleanElapsed := time.Since(cleanStart)
			fmt.Println(cleanElapsed)
			tick = 0
		}

	}

}
