package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
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
	Output chan Response
}

// create http handler for HandleFunc that writes Requests to output
func makeHandler(output chan Request) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// Create the output channel to get the result back from the cache of
		// the redis subsystems
		request := Request{
			Key:    strings.TrimPrefix(req.URL.Path, "/"),
			Output: make(chan Response),
		}

		// send request to cache subsystem
		output <- request

		// wait for the response to come back
		response := <-request.Output
		if response.Err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			io.WriteString(w, response.Err.Error())
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(response.Result)
		}
	}
}

func main() {
	var cacheQueue = make(chan Request, 1234) // FIXME use queue size
	var redisQueue = make(chan Request, 1234) // FIXME
	_ = redisQueue

	http.HandleFunc("/", makeHandler(cacheQueue))

	go func() {
		fmt.Printf("%#v\n", <-cacheQueue)
	}()

	log.Fatal(http.ListenAndServe(":8080", nil))
}
