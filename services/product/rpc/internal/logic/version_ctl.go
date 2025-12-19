package logic

import (
	"context"
	"fmt"
	"strconv"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

func GetCategoryVersion(ctx context.Context, category string, rds *redis.Redis) int64 {
	if category == "" {
		return GetGlobalVersion(ctx, rds)
	}

	cacheKey := fmt.Sprintf("product:CategoryVersion:%s", category)
	version, err := rds.GetCtx(ctx, cacheKey)
	if err != nil || version == "" {
		rds.SetexCtx(ctx, cacheKey, "1", 86400*365) // categoryKey存一年
		return 1
	}

	v, _ := strconv.ParseInt(version, 10, 64)
	return v
}

func IncCategoryVersion(ctx context.Context, category string, rds *redis.Redis) error {
	cacheKey := fmt.Sprintf("product:CategoryVersion:%s", category)
	_, err := rds.IncrCtx(ctx, cacheKey)
	return err
}

func GetGlobalVersion(ctx context.Context, rds *redis.Redis) int64 {
	cacheKey := "product:GlobalVersion"
	version, err := rds.GetCtx(ctx, cacheKey)
	if err != nil || version == "" {
		rds.SetexCtx(ctx, cacheKey, "1", 86400*365)
		return 1
	}

	v, _ := strconv.ParseInt(version, 10, 64)
	return v
}

func IncGlobalVersion(ctx context.Context, rds *redis.Redis) error {
	cacheKey := "product:GlobalVersion"
	_, err := rds.IncrCtx(ctx, cacheKey)
	return err
}
