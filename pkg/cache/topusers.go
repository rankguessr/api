package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/rankguessr/api/pkg/domain"
	"github.com/redis/go-redis/v9"
)

func buildTopUsersKey(limit, page int) string {
	return fmt.Sprintf("topusers:%d:%d", limit, page)
}

func SetTopUsers(rdb *redis.Client, ctx context.Context, users domain.Paged[domain.UserExtended], limit, page int) error {
	return SetJSON(rdb, ctx, buildTopUsersKey(limit, page), users, 30*time.Second)
}

func GetTopUsers(rdb *redis.Client, ctx context.Context, limit, page int) (domain.Paged[domain.UserExtended], error) {
	return GetJSON[domain.Paged[domain.UserExtended]](rdb, ctx, buildTopUsersKey(limit, page))
}
