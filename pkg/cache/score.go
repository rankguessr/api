package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/rankguessr/api/pkg/osuapi"
	"github.com/redis/go-redis/v9"
)

func scoreCacheKey(scoreId int) string {
	return fmt.Sprintf("score:%d", scoreId)
}

func SetScore(rdb *redis.Client, ctx context.Context, score osuapi.Score) error {
	return SetJSON(rdb, ctx, scoreCacheKey(score.ID), score, 5*time.Minute)
}

func GetScore(rdb *redis.Client, ctx context.Context, scoreId int) (osuapi.Score, error) {
	return GetJSON[osuapi.Score](rdb, ctx, scoreCacheKey(scoreId))
}
