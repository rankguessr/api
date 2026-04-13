package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

func SetJSON(rdb *redis.Client, ctx context.Context, key string, val any, ttl time.Duration) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}

	return rdb.Set(ctx, key, data, ttl).Err()
}

func GetJSON[T any](rdb *redis.Client, ctx context.Context, key string) (T, error) {
	var result T
	data, err := rdb.Get(ctx, key).Bytes()
	if err != nil {
		return result, err
	}

	return result, json.Unmarshal(data, &result)
}
