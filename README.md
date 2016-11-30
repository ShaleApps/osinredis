## Redis storage for osin

Provides Redis-based storage for [osin](https://github.com/RangelReale/osin) and is based on [redigo](https://github.com/garyburd/redigo).

### Installation

```
$ go get github.com/ShaleApps/osinredis
```

### Running tests

The tests are integration tests that require a running Redis instance. Each test runs `FLUSHALL` on the server so if you don't want to run against the default local address (`:6379`), provide a different address via a `REDIS_ADDR` environment variable
```
$ REDIS_ADDR=:1234 go test ./...
```

### Usage

Example:

```go
import (
    "github.com/RangelReale/osin"
    "github.com/ShaleApps/osinredis"
	"github.com/garyburd/redigo/redis"
)

func main() {
	pool = &redis.Pool{
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", ":6379")
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}

	storage := osinredis.New(pool, "prefix")
	server := osin.NewServer(osin.NewServerConfig(), storage)
}
```
