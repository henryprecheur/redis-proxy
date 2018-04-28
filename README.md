Redis Proxy
===========

Redis Proxy is a caching HTTP server for the Redis in-memory database.

Requirements & build
--------------------

Build & test requirements: Docker, docker-compose with support for format
version 3 and above, and make.

To build the docker image for the service:

    $ make build

Integration testing is done via `make test` inside a docker environment created
via docker-compose:

    $ make test

You can also build the program locally with:

    $ go build

Run it
------

Here’s the quick and easy way to get redis-proxy running:

    $ redis-proxy -redis redis.example.com:6379 -http :8080

‘redis.example.com:6379’ is the address of the backing Redis server to use, and
‘:8080’ tells redis-proxy to listen on port 8080 on all interfaces. You can
specify the address redis-proxy listens on, for example to listen only on the
local interface you’d pass ‘localhost:8080’.

You can also configure the cache via the -expire and -capacity options. For
example to expire keys after 2 minutes of non-use or when we have more than 1000
keys in the cache.

    $ redis-proxy -expire 2m -capacity 1000

redis-proxy can handle multiple HTTP client concurrently, to limit the number
of in-flight request to Redis there’s the -redis-ops-limit parameter.

To learn more about redis-proxy’s options:

    $ redis-proxy -h

You can then query the proxy like this:

    $ curl http://localhost:8080/example_key
    some value

redis-proxy doesn’t have a configuration file: all options are passed on the
command-line.

Architecture
------------

Dependencies are managed with [glide][] to guarantee reproducible builds. The
only depency for now is [redisgo][], it is shipped as part of the repository in
the vendor/ directory.

[glide]: https://glide.sh/

redis-proxy has three parts: HTTP front-end, Redis backend, and Cache

The front-end uses the [net/http](https://godoc.org/net/http) package. Every
time a new client connects a new goroutine is spawned that talks to the Cache
via a buffered channel.

The cache handles requests from the HTTP front-end. It’s two goroutine that
runs in the background, the main goroutine reads its input from a buffered
channel that the front-end writes to. For each request, the cache checks if the
key is present in its local cache and return its corresponding value. If the key
is absent, the cache forward the request to the Redis backend via a buffered
channel.

There’s a background goroutine to trigger time-based key expiries.

The Redis backend has a two goroutines running in the background. They read its
input from a buffered channel, asyncronously forward the command to Redis. Once
the response comes back: the Redis backend forwards it back to the cache
subsystem that stores the value, and forward it to the HTTP goroutine to
serve the result back to the HTTP client.

Cache operations complexity
---------------------------

The cache is based on [sync.Map](https://godoc.org/sync#Map), according to the
documentation:

> Loads, stores, and deletes run in amortized constant time.

The system tracks the next entries to expire via a priority queue (see
priority_queue.go).

Time-based key expiries are triggered by a timer, it runs every second. Once the
timer is triggered the system removes the corresponding key from the cache and
scans the cache to find when the next key to be expired until there's no key
that needs to be removed. It uses a [heap][] as a priority list to reduce the
complexity of the scan: A linear scan would be O(n), but by using a heap we
reduce the complexity to O(log(n)).

[heap]: https://godoc.org/container/heap

Size-cap key expiries are triggered when a new key gets queried. Similarly to
time-based expiries the proxy removes the oldest key and it scans
the cache to find the next key to expire. Like time-based expiries it is a O(n)
operation that could be reduced to O(log(n)) with a heap to scan the cache.

Uncached queries have a complexity of O(log(n)) because we need to insert the
new value into the priority queue.

Cached queries are constant time operations.

Known-issues
------------

I didn’t have enough time to work on this:

- Redis client protocol: the service speaks only HTTP, I missed that requirement
  when I got started. Adding support for it requires extra-time.
- Error handling is poor: if the Redis backing server goes down for example, the
  redis-proxy won’t reconnect or retry failed queries: it will just exit.

Time-tracking
-------------

1. Design & write documentation, includes architecture & complexity of caches
   operations: 1 hour estimated, **actual time 1h45**
2. Setup build/test environment with docker-compose: 30 minutes estimated
   **actual time 10 minutes**
3. Write service that executes Redis commands sequentially: 30 minutes **actual
   time 2 hours**
4. Write integration tests: key caching, retrieval, and expiry all tested: 45
  minutes **actual time 30 minutes**
5. Handle multiple concurrent access to proxy with a Redis pipeline: 30 minutes
   **included in step #2**
