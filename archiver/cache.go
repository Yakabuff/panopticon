package main

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type redisClient struct {
	conn *redis.Client
	ctx  context.Context
}

// Persistent cache for
// 1) image hashes to prevent downloading again (lru)
// 2) post/thread id to prevent making unncessary db writes (tid -> [pids])
func redisInit() (*redisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	var ctx = context.Background()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}
	return &redisClient{conn: rdb, ctx: ctx}, nil
}

func (r *redisClient) insertPid(board string, tid string, pid string) error {
	err := r.conn.SAdd(r.ctx, board+tid, pid).Err()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
func (r *redisClient) checkThreadExist(board string, tid string) (bool, error) {
	val, err := r.conn.Exists(r.ctx, board+tid).Result()
	if err != nil {
		return false, err
	}
	if val > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (r *redisClient) deleteTid(board string, tid string) error {
	err := r.conn.Del(r.ctx, board+tid).Err()
	if err != nil {
		fmt.Println(err)
		return err
	}
	return err
}

func (r *redisClient) checkPidExist(board string, tid string, pid string) (bool, error) {
	res, err := r.conn.SIsMember(r.ctx, board+tid, pid).Result()
	if err != nil {
		return false, err
	}
	return res, nil
}

func (r *redisClient) insertHash(hash string) error {
	size, err := r.conn.ZCard(r.ctx, "IMAGE_HASHES").Result()
	if err != nil {
		return err
	}
	if size > 100000 {
		r.conn.ZPopMin(r.ctx, "IMAGE_HASHES", 50000)
	}
	r.conn.ZIncrBy(r.ctx, "IMAGE_HASHES", 1, hash)
	return nil
}

func (r *redisClient) checkHashExists(hash string) (bool, error) {
	_, err := r.conn.ZScore(r.ctx, "IMAGE_HASHES", hash).Result()

	if err == redis.Nil {
		// key does not exist
		return false, nil
	} else if err == nil {
		// key exists
		return true, nil
	} else {
		// actual error
		return false, err
	}
}
