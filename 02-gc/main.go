package main

import (
	"fmt"
	"runtime"
	"time"
)

type Node struct {
	ID   int
	Next *Node
}

var sink []*Node

func main() {
	start := time.Now()
	for i := 0; i < 10_000_000; i++ {
		n := &Node{ID: i}
		if i%100 == 0 {
			sink = append(sink, n)
		}
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("time=%v numGC=%d heapAlloc=%dMB totalAlloc=%dMB pauseTotal=%v\n",
		time.Since(start), m.NumGC, m.HeapAlloc>>20, m.TotalAlloc>>20,
		time.Duration(m.PauseTotalNs))
}
