package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/rankguessr/api/pkg/domain"
	"github.com/redis/go-redis/v9"
)

func buildStatsKey(limit, page int) string {
	return fmt.Sprintf("stats:%d:%d", limit, page)
}

func SetStats(rdb *redis.Client, ctx context.Context, stats domain.Stats, limit, page int) error {
	return SetJSON(rdb, ctx, buildStatsKey(limit, page), stats, 3*time.Minute)
}

func GetStats(rdb *redis.Client, ctx context.Context, limit, page int) (domain.Stats, error) {
	return GetJSON[domain.Stats](rdb, ctx, buildStatsKey(limit, page))
}
