package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// NewClient membuat dan mengembalikan koneksi Redis baru.
func NewClient(redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	// Ping server Redis untuk memeriksa koneksi
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}
