package main

import (
	"arduino-serial/errors"
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func InitRedis(dsn string, maxRetry int) (*redis.Client, error) {
	var rdb *redis.Client
	sleepDuration := 1
	retry := 0

	opt, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed on parsing redis dsn")
	}

	for retry < maxRetry {
		rdb = redis.NewClient(opt)
		ctx := context.Background()

		err := rdb.Ping(ctx).Err()
		if err == nil {
			break
		}

		log.Err(err).Msgf("failed on creating connection on redis, retrying in %d second", sleepDuration)
		time.Sleep(time.Duration(sleepDuration * int(time.Second)))
		sleepDuration++
		retry++
	}

	if rdb == nil {
		return nil, errors.New("failed on creating connection on redis")
	}

	return rdb, nil
}
