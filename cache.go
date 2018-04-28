package main

import (
	"container/heap"
	"log"
	"sync"
	"time"
)

// Value stored in our map and priority queue
type Item struct {
	Key    string
	Value  []byte
	Expiry time.Time
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

type Cache struct {
	m           sync.Map
	expiryQueue *PriorityQueue
	// ensure we don't have 2 goroutine removing from our cache at the same time
	expireLock sync.Mutex

	// configuration
	maxSize    int
	timeToLive time.Duration

	// the cache's own input
	Input chan *Request
	// the redis backend input for when keys aren't cached
	redisChan chan *Request
}

func NewCache(redisChan chan *Request, maxSize int, timeToLive time.Duration) *Cache {
	c := new(Cache)
	c.expiryQueue = NewPriorityQueue()

	c.maxSize = maxSize
	c.timeToLive = timeToLive
	c.redisChan = redisChan
	c.Input = make(chan *Request, 64) // FIXME hardcoded value
	return c
}

func (c *Cache) expiryTime() time.Time {
	return time.Now().Add(c.timeToLive)
}

func (c *Cache) addKey(key string, value []byte) {
	item := &Item{Key: key, Value: value, Expiry: c.expiryTime()}
	c.m.Store(key, item)

	// update the priority queue in the background
	go func() {
		c.expireLock.Lock()
		heap.Push(c.expiryQueue, item)
		c.expireKeys()
		c.expireLock.Unlock()
	}()
}

func (c *Cache) expireKeys() {
	log.Println("cache: remove expired keys")
	// Remove the extra entries in the expiry queue if needed
	if x := c.expiryQueue.Len() - c.maxSize; x > 0 {
		for x > 0 {
			item := heap.Pop(c.expiryQueue).(*Item)
			c.m.Delete(item.Key)
			x -= 1
		}
	}

	// Remove expired entries if needed
	now := time.Now()
	for c.expiryQueue.Oldest().After(now) {
		item := heap.Pop(c.expiryQueue)
		c.m.Delete(item.Key)
	}

	c.expireLock.Unlock()
}

// garbage collect expired keys every period
func (c *Cache) backgroundExpiry(period time.Duration) {
	ticker = time.NewTicker(period)
	for {
		select {
		case <-ticker.C:
			c.expireKeys()
		}
	}
}

func (c *Cache) Process() {
	go c.backgroundExpiry(time.Second)  // FIXME hardcoded value, should be config

	for r := range c.Input {
		val, ok = c.m.Load(r.Key)
		if !ok {
			// key not found, we forward the request to the Redis backend and
			// wait for the result in a separate goroutine
			go func() {
				c := make(chan Response)
				c.redisQueue <- Request{Key: r.Key, Output: c}
				resp := <-c // wait for the Redis backend's response
				// Add the response to our cache
				if resp.Err != nil {
					c.AddKey(r.Key, resp.Result)
				}
			}
		}
		// Forward response to HTTP front-end
		r.Output <- resp
	}
}
