package main

import (
	"log"

	"github.com/gomodule/redigo/redis"
)

type Database struct {
	Input chan *Request
}

func NewDatabase(addr string, maxRequest int) (*Database, error) {
	db := new(Database)
	db.Input = make(chan *Request)

	if conn, err := redis.Dial("tcp", addr); err != nil {
		return nil, err
	} else {
		go db.run(maxRequest, conn)
	}

	return db, nil
}

// receiver processes response from the Redis server. It gets the in-flight
// requests via the input channel, reads the result from Redis, and forward the
// result to the Response channel
func (d *Database) receiver(input chan *Request, conn redis.Conn) {
	for {
		// read next in-flight request
		r := <-input
		// now read the corresonding result from Redis
		x, err := conn.Receive()
		log.Println("redis: got response for", r.Key)
		r.Output <- &Response{Result: x.([]byte), Err: err}
	}
}

// main loop for the database
func (d *Database) run(maxRequest int, conn redis.Conn) {
	// We limit the number of in-flight requests to Redis
	receiverChan := make(chan *Request, maxRequest)
	go d.receiver(receiverChan, conn)

	for {
		// Get all the requests we can from the queue and exec them all at once
		// via the pipeline

		// wait for a request to come in
		r := <-d.Input
		log.Println("redis: sending requests for", r.Key)
		receiverChan <- r // send request to receiver for further processing
		conn.Send("GET", r.Key)
		conn.Flush()
	}
}
