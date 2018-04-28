package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

//
// HTTP Front-end
//

// Response for the HTTP handler, can be sent by either the Cache or Redis subsystem
type Response struct {
	Result []byte
	Err    error
}

type Request struct {
	Key    string
	Output chan *Response
}

// create http handler for HandleFunc that writes Requests to output
func makeHandler(output chan *Request) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		key := strings.TrimPrefix(req.URL.Path, "/")
		// output channel to get the result back from the cache or
		// the redis subsystems
		o := make(chan *Response)
		request := Request{Key: key, Output: o}
		log.Println("http: Got request for", key)

		// send request to cache subsystem
		output <- &request
		log.Println("http: Sent request for", key)

		log.Println("http", request.Output)
		// wait for the response to come back
		response := <-request.Output
		log.Println("http: Got response for", key)
		if response.Err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, response.Err.Error())
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(response.Result)
		}
	}
}

var (
	httpAddr     string
	dbAddr       string
	maxCacheSize uint
	expiry       time.Duration
	maxRequest   uint
)

func init() {
	flag.StringVar(&httpAddr, "http", ":8080", "address:port to listen on")
	flag.StringVar(&dbAddr, "redis", "localhost:6379", "address:port to the Redis server")
	flag.UintVar(&maxCacheSize, "capacity", 1000, "maximum number of item to keep in cache")
	flag.DurationVar(&expiry, "expire-after", time.Minute, "how long the cache will keep values")
	flag.UintVar(&maxRequest, "redis-ops-limit", 10, "maximum number of concurrent Redis operation")
}

func main() {
	flag.Parse()

	log.Println(httpAddr)
	// redis back-end
	database, err := NewDatabase(dbAddr, int(maxRequest))
	if err != nil {
		log.Fatal(err)
	}
	// cache back-end
	cache := NewCache(database.Input, int(maxCacheSize), expiry)
	go cache.Process()
	// http front-end
	http.HandleFunc("/", makeHandler(cache.Input))

	log.Fatal(http.ListenAndServe(httpAddr, nil))
}
