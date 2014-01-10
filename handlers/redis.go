package handlers

import (
	"github.com/garyburd/redigo/redis"
	"github.com/wurkhappy/WH-Config"
	"time"
)

var redisPool *redis.Pool
var whName string

func Setup(production bool) {
	if production {
		whName = "reply"
	} else {
		whName = "test"
	}
	var password string
	var network string = "tcp"
	var address string = config.WebAppRedis

	redisPool = &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial(network, address)
			if err != nil {
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
