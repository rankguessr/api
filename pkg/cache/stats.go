package cache

import (
	"context"
	"time"

	"github.com/rankguessr/api/pkg/domain"
	"github.com/redis/go-redis/v9"
)

func buildStatsKey() string {
	return "stats"
}

func SetStats(rdb *redis.Client, ctx context.Context, stats domain.Stats) error {
	return SetJSON(rdb, ctx, buildStatsKey(), stats, 30*time.Second)
}

func GetStats(rdb *redis.Client, ctx context.Context) (domain.Stats, error) {
	return GetJSON[domain.Stats](rdb, ctx, buildStatsKey())
}
