package kumnats

import (
	"time"

	redigo "github.com/gomodule/redigo/redis"
)

type Option func(*Options) error

type Options struct {
	redisConn                             *redigo.Pool
	failedMessagesRedisKey                string
	deadMessagesRedisKey                  string
	reconnectInterval                     time.Duration
	failedMessagePublishIntervalInSeconds uint64
	logger                                Logger
}

func WithRedis(conn *redigo.Pool) Option {
	return func(opt *Options) error {
		opt.redisConn = conn
		return nil
	}
}

func WithFailedMessageRedisKey(key string) Option {
	return func(opt *Options) error {
		opt.failedMessagesRedisKey = key
		return nil
	}
}

func WithDeadMessageRedisKey(key string) Option {
	return func(opt *Options) error {
		opt.deadMessagesRedisKey = key
		return nil
	}
}

func WithReconnectInterval(duration time.Duration) Option {
	return func(opt *Options) error {
		opt.reconnectInterval = duration
		return nil
	}
}

func WithFailedMessagePublishInterval(seconds uint64) Option {
	return func(opt *Options) error {
		opt.failedMessagePublishIntervalInSeconds = seconds
		return nil
	}
}

func WithLogger(logger Logger) Option {
	return func(opt *Options) error {
		opt.logger = logger
		return nil
	}
}
