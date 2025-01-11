package main

import (
	"context"
	"fmt"
)

// var redisCtx = context.Background()

type Store interface {
	GetAll() map[string]string
	Delete()
	Save()
	Get()
}

func getRedisKey(sessionID string) string {
	return fmt.Sprintf("client:%s:urls", sessionID)
}

func (a *app) storeTLInRedis() {}

func (a *app) deleteTLFromRedis() {}

func (a *app) getAllTLsFromRedis(ctx context.Context, rkey string) map[string]string {
	urls, err := a.rdb.HGetAll(ctx, rkey).Result()
	if err != nil {
		a.logger.Info("failed to retrieve URLs from Redis", "error", err)
		return nil
	}
	return urls
}

func (a *app) getTLFromRedis(ctx context.Context, tinylink, rkey string) (string, error) {
	redisKey := fmt.Sprintf("client:%s:urls", rkey)
	val, err := a.rdb.HGet(ctx, redisKey, tinylink).Result()
	if err != nil {
		return "", err
	}
	return val, err
}
