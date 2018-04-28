package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/gomodule/redigo/redis"
)

func redisSet(t *testing.T, conn redis.Conn, key, value interface{}) {
	if _, err := conn.Do("SET", key, value); err != nil {
		t.Fatalf("SET %v %v failed: %s", key, value, err)
	}

}

func TestCache(t *testing.T) {
	time.Sleep(10) // wait for stuff to come up
	conn, err := redis.Dial("tcp", "localhost:6379")
	if err != nil {
		t.FailNow()
	}
	redisSet(t, conn, "foo", "a")

	resp, err := http.Get("http://localhost:8080/foo")
	if err != nil {
		t.Fatal("err")
	}
	b, err := ioutil.ReadAll(resp.Body)
	if bytes.Compare(b, []byte{'a'}) != 0 {
		t.Fatal("body")
	}
}

func TestExpiry(t *testing.T) {

}
