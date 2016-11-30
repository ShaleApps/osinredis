/*
A Redis storage backend for osin.

Installation:

	go get github.com/ShaleApps/osinredis

Usage:

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

*/
package osinredis
