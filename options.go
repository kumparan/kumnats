package kumnats

import (
	"time"

	redigo "github.com/gomodule/redigo/redis"
)

// Option :nodoc:
type Option func(*Options) error

// Options :nodoc:
type Options struct {
	redisConn                             *redigo.Pool
	failedMessagesRedisKey                string
	deadMessagesRedisKey                  string
	reconnectInterval                     time.Duration
	failedMessagePublishIntervalInSeconds uint64
	logger                                Logger
}

// WithRedis :nodoc:
func WithRedis(conn *redigo.Pool) Option {
	return func(opt *Options) error {
		opt.redisConn = conn
		return nil
	}
}

// WithFailedMessageRedisKey :nodoc:
func WithFailedMessageRedisKey(key string) Option {
	return func(opt *Options) error {
		opt.failedMessagesRedisKey = key
		return nil
	}
}

// WithDeadMessageRedisKey :nodoc:
func WithDeadMessageRedisKey(key string) Option {
	return func(opt *Options) error {
		opt.deadMessagesRedisKey = key
		return nil
	}
}

// WithReconnectInterval :nodoc:
func WithReconnectInterval(duration time.Duration) Option {
	return func(opt *Options) error {
		opt.reconnectInterval = duration
		return nil
	}
}

// WithFailedMessagePublishInterval :nodoc:
func WithFailedMessagePublishInterval(seconds uint64) Option {
	return func(opt *Options) error {
		opt.failedMessagePublishIntervalInSeconds = seconds
		return nil
	}
}

// WithLogger :nodoc:
func WithLogger(logger Logger) Option {
	return func(opt *Options) error {
		opt.logger = logger
		return nil
	}
}
