package socks5t

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

const (
	DefaultSendChan = "send"
	DefaultRecvChan = "recv"
)

type RedisStreamer struct {
	pool      *redis.Pool
	subChan   string
	handler   func(*redis.Message) error
	logger    *logrus.Entry
	running   atomic.Bool
	listening atomic.Bool
	cancel    context.CancelFunc
}

func NewRedisStreamer(pool *redis.Pool, subChan string) *RedisStreamer {
	return &RedisStreamer{
		pool:    pool,
		subChan: subChan,
		logger:  logrus.WithContext(context.TODO()),
	}
}

func (rs *RedisStreamer) SetHandler(handler func(*redis.Message) error) {
	rs.handler = handler
}

func (rs *RedisStreamer) Run(ctx context.Context) {
	if !rs.running.CompareAndSwap(false, true) {
		rs.logger.Warningf("running")
		return
	}

	err := rs.run(ctx)
	if err != nil {
		rs.logger.Errorf("run err: %s", err)
	}
}

func (rs *RedisStreamer) handleMessage(v *redis.Message) {
	if v.Channel == rs.subChan {
		err := rs.handler(v)
		if err != nil {
			rs.logger.Errorf("handle msg [%s](%s) failed: %s", v.Channel, v.Data, err)
		}
	} else {
		rs.logger.Warningf("unexpected ch: %s", v.Channel)
	}
}

func (rs *RedisStreamer) run(ctx context.Context) error {
	ctx, rs.cancel = context.WithCancel(ctx)

	conn := rs.pool.Get()
	defer conn.Close()

START:
	psconn := redis.PubSubConn{Conn: conn}

	rs.logger.Infof("Listen: %v", rs.subChan)

	err := psconn.Subscribe(rs.subChan)
	if err != nil {
		return err
	}
	for rs.running.Load() {
		switch v := psconn.ReceiveContext(ctx).(type) {
		case redis.Message:
			go rs.handleMessage(&v)
		case redis.Subscription:
			if v.Channel == rs.subChan {
				rs.logger.Infof("subscription %#v", v)
				if v.Kind == "subscribe" {
					rs.listening.Store(true)
				} else if v.Kind == "unsubscribe" {
					rs.listening.Store(false)
				}
			} else {
				rs.logger.Warningf("unexpected subscription %#v", v)
			}
		case error:

			if v.Error() == "context canceled" {
				rs.logger.Warningf("context canceled, stop")
				rs.listening.Store(false)
				rs.running.Store(false)
			} else {
				rs.logger.Errorf("rds recieve err: %s", v)
				conn.Close()
				rs.listening.Store(false)
				conn = rs.pool.Get()
				goto START
			}
		default:
			rs.logger.Warningf("Unknown: %#v", v)
		}

	}

	rs.listening.CompareAndSwap(true, false)
	return err
}

func (rs *RedisStreamer) Listening() bool {
	return rs.running.Load() && rs.listening.Load()
}

func (rs *RedisStreamer) Running() bool {
	return rs.running.Load()
}

func (rs *RedisStreamer) Pub(chname string, data interface{}) (r int64, err error) {

	conn := rs.pool.Get()
	defer conn.Close()
	r, err = redis.Int64(conn.Do("PUBLISH", chname, data))
	return

}

func (rs *RedisStreamer) Stop() {
	if !rs.running.CompareAndSwap(true, false) {
		return
	}
	rs.cancel()
}

func (rs *RedisStreamer) Join() {
	for rs.listening.Load() || rs.running.Load() {

	}
}

func NewRedisPool(DSN string, maxIdle int, maxActive int) (*redis.Pool, error) {

	redisPool := &redis.Pool{
		MaxIdle:     maxIdle,
		MaxActive:   maxActive,
		IdleTimeout: time.Minute,
		Dial: func() (redis.Conn, error) {
			c, err := redis.DialURL(DSN)
			if err != nil {
				return nil, err
			}
			return c, err
		},
	}
	conn := redisPool.Get()
	defer func(conn redis.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Printf("close err: %s", err)
		}
	}(conn)
	if _, err := redis.String(conn.Do("PING")); err != nil {
		return nil, err
	}
	return redisPool, nil
}
