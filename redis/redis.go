package redis

import (
	"menteslibres.net/gosexy/redis"
)

const (
	hostname = "127.0.0.1"
	port     = 6379
)

func New() (*redis.Client, error) {
	r := redis.New()
	err := r.Connect(hostname, port)
	if err != nil {
		return nil, err
	}
	return r, nil
}
