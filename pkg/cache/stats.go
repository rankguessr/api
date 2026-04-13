package cache

import (
	"context"
	"time"

	"github.com/rankguessr/api/pkg/domain"
	"github.com/redis/go-redis/v9"
)

func SetStats(rdb *redis.Client, ctx context.Context, stats domain.Stats) error {
	return SetJSON(rdb, ctx, "stats", stats, 3*time.Minute)
}

func GetStats(rdb *redis.Client, ctx context.Context) (domain.Stats, error) {
	return GetJSON[domain.Stats](rdb, ctx, "stats")
}
