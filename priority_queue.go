// This code was largely copy-pasted from the container/heap examples:
// https://play.golang.org/p/nqe-zZ4o58R
package main

import (
	"container/heap"
	"time"
)

// Entry for our priority list
type Item struct {
	Key    string
	Expiry *time.Time // pointer to the value stored in the map
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// implements heap.Interface and holds Items.
type PriorityQueue []*Item

func NewPriorityQueue() *PriorityQueue {
	x := new(PriorityQueue)
	heap.Init(x)
	return x
}

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].expiry.Before(*pq[j].expiry)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// Return the time of the oldest entry in the priority queue
func (pq *PriorityQueue) Oldest() *time.Time {
	return pq[len(pq)-1].expiry
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, expiry *time.Time) {
	item.expiry = expiry
	heap.Fix(pq, item.index)
}
