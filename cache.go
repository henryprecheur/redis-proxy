package main

import (
	"container/heap"
	"sync"
	"time"
)

// Value stored in our map
type Value struct {
	Value  []byte
	Expiry time.Time
}

type Cache struct {
	m           sync.Map
	expiryQueue *PriorityQueue
	// ensure we don't have 2 goroutine removing from our cache at the same time
	expireLock sync.Mutex

	// configuration
	maxSize    uint
	timeToLive time.Duration

	// the cache's own input
	Input chan Request
	// the redis backend input when keys aren't cached
	redisInput chan Request
}

func NewCache(redisQueue chan Request, maxSize uint, timeToLive time.Duration) *Cache {
	c := new(Cache)
	c.expiryQueue = NewPriorityQueue()

	c.maxSize = maxSize
	c.timeToLive = timeToLive
	c.redisInput = redisQueue
	c.Input = make(chan Request, 64) // FIXME hardcoded value
	return c
}

func (c *Cache) expiryTime() time.Time {
	return time.Now().Add(c.timeToLive)
}

func (c *Cache) AddKey(key string, value []byte) {
	val := Value{Value: value, Expiry: c.expiryTime()}
	c.m.Store(key, val)

	// update the priority queue in the background
	go func() {
		e := Item{key: key, expiry: &(val.Expiry)}
		c.expireLock.Lock()
		heap.Push(c.expiryQueue, e)
		c.expireLock.Unlock()

		c.expireKeys()
	}()
}

func (c *Cache) expireKeys() {
	c.expireLock.Lock()

	// Remove the extra entries in the expiry queue if needed
	if x := len(c.expiryList) - c.maxSize; x > 0 {
		for x {
			item := heap.Pop(c.expiryQueue)
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

func (c *Cache) Process(expiryPeriod time.Duration) {
	go c.backgroundExpiry(expiryPeriod)

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
