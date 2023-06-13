package main

import (
	"context"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client

func RedisInit() {
	RDB = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Delete previous keys (may be outdated)
	ctx := context.Background()
	RDB.FlushDB(ctx)
	Log.Info("Flushed Redis")
}

func set(ctx context.Context, key string, value string) error {
	_, err := RDB.Set(ctx, key, value, 0).Result()
	if err != nil {
		Log.WithError(err).Error("Failed to set " + key + ":" + value)
	}
	return err
}

func get(ctx context.Context, key string) (string, error) {
	value, err := RDB.Get(ctx, key).Result()
	if err != nil {
		Log.WithError(err).Error("Failed to get " + key)
	}
	return value, err
}
