Redis Proxy
===========

Redis Proxy is a caching HTTP server for the Redis in-memory database.

Requirements & build
--------------------

Golang 1.9+ is required to build Redis proxy. Redis proxy also depends on the
[redis.v5](https://godoc.org/gopkg.in/redis.v5) and the
[pflag](https://godoc.org/github.com/spf13/pflag) packages, and it uses
[glide][] to manage dependencies. To fetch the dependencies and build the
program locally you can use:

    $ make build

It also requires support for docker-compose format version 3 or above (FIXME
which version of docker-compose is that?).

Integration testing is done via `make test` inside a docker environment created
via docker-compose.

    $ make test

Run it
------

Here’s the quick and easy way to get redis-proxy running:

    $ redis-proxy redis.example.com:6379 :8080

‘redis.example.com:6379’ is the address of the backing Redis server to use, and
‘:8080’ tells redis-proxy to listen on port 8080 on all interfaces. You can
specify the address redis-proxy listens on, for example to listen only on the
local interface you’d pass ‘localhost:8080’.

You can also configure the cache via the --expire and --capacity options. For
example to expire keys after 2 minutes of non-use or when we have more than 1000
keys in the cache.

    $ redis-proxy redishost:6379 :8080 --expire 2m --capacity 1000

redis-proxy can handle multiple HTTP client concurrentely, to limit the number
of in-flight request to Redis there’s the --queue-limit parameter.

To learn more about redis-proxy’s options:

    $ redis-proxy --help

You can then query the proxy like this:

    $ curl http://localhost:8080/example_key
    some value

redis-proxy doesn’t have a configuration file: all options are passed on the
command-line.

Architecture
------------

Dependencies are managed with [glide][] to guarantee reproducible builds.

[glide]: https://glide.sh/

redis-proxy has three main parts: HTTP front-end, Redis backend, and Cache

The front-end relies on the [net/http](https://godoc.org/net/http) package to
handle HTTP connections. Every time a new client connects a new goroutine is
spawned that talks to the Cache via a buffered channel.

The cache handles requests from the HTTP front-end. It’s two goroutine that
runs in the background, the main goroutine reads its input from a buffered
channel that the front-end writes-to. For each request, the cache checks if the
key is present in its local cache and return its corresponding value. If the key
is absent, the cache forward the request to the Redis backend via a buffered
channel. There’s a second goroutine to trigger time-based key expiries.

The Redis backend has a single goroutine running in the background. It reads its
input from a buffered channel. The Redis backend uses
[pipelines](https://redis.io/topics/pipelining) to send the GET command
asynchronously to Redis and spawn a new goroutine to transmit the result from
Redis synchronously to the caller once it comes back.

Cache operations complexity
---------------------------

The cache is based on [sync.Map][], according to the documentation:

> Loads, stores, and deletes run in amortized constant time.

Time-based key expiries are triggered by a timer. The timer tracks the time to
live for the next key to be expired. Once the timer is triggered the system
removes the corresponding key from the cache and scans the cache to find when
the next key to be expired and reset the timer with its expiry time. Because it
scans the entire cache every time it removes a key this operation’s complexity
is O(n) where n is the size of the cache. We could reduce the cost of scanning
the cache by tracking keys’ expiry time in a [heap][], this would reduce the
complexity of the scan to O(log(n)).

[heap]: https://godoc.org/container/heap

Size-cap key expiries are triggered when a new key gets queried. Similarly to
time-based expiries the proxy removes the oldest key and it scans
the cache to find the next key to expire. Like time-based expiries it is a O(n)
operation that could be reduced to O(log(n)) with a heap to scan the cache.

Time-tracking
-------------

1. Design & write documentation, includes architecture & complexity of caches
   operations: 1 hour estimated, **actual time 1h30**
2. Setup build/test environment with docker-compose: 30 minutes estimated
   **actual time 10 minutes**
3. Write service that executes Redis commands sequentially: 30 minutes **actual
   time 1 hour**
4. Write integration tests: key caching, retrieval, and expiry all tested: 45
  minutes **acutal time ???***
5. Handle multiple concurrent access to proxy with a Redis pipeline: 30 minutes
   **acutal time ???***
