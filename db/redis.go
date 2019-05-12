package db

import (
        "context"
        "time"

        "github.com/gomodule/redigo/redis"
        "github.com/pkg/errors"
)

var Redis *redis.Pool

func InitRedis(addr, password string, db, maxIdle, maxActive int, dialTimeout time.Duration) {
        var options = []redis.DialOption{
                redis.DialDatabase(db),
                redis.DialPassword(password),
        }
        if dialTimeout > 0 {
                options = append(options, redis.DialConnectTimeout(dialTimeout))
        }

        pool := &redis.Pool{
                Dial: func() (conn redis.Conn, e error) {
                        c, err := redis.Dial("tcp", addr, options...)
                        if err != nil {
                                return nil, err
                        }
                        return c, nil
                },
                TestOnBorrow: func(c redis.Conn, t time.Time) error {
                        _, err := c.Do("PING")
                        return err
                },
                MaxIdle:   maxIdle,
                MaxActive: maxActive,
        }

        if pool.MaxActive > 0 {
                pool.Wait = true
        }
        Redis = pool
}

// DO sends a command to the server and returns the receive reply.
func getConn(ctx context.Context) (redis.Conn, error) {
        conn, err := Redis.GetContext(ctx)
        if err != nil {
                return nil, errors.Wrap(err, "redis error.")
        }
        if err := conn.Err(); err != nil {
                return nil, errors.Wrap(err, "redis error.")
        }
        return conn, nil
}

func Del(ctx context.Context, keys ...interface{}) error {
        conn, err := getConn(ctx)
        if err != nil {
                return err
        }
        defer conn.Close()
        _, err = conn.Do("del", keys...)
        if err != nil {
                return errors.Wrap(err, "")
        }
        return nil
}

func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
        conn, err := getConn(ctx)
        if err != nil {
                return err
        }
        defer conn.Close()

        if expiration < 0 {
                if _, err := conn.Do("set", key, value); err != nil {
                        return errors.Wrap(err, "")
                }
        }
        if expiration < time.Second && expiration > time.Millisecond {
                if _, err := conn.Do("set", key, value, "px", int64(expiration/time.Millisecond)); err != nil {
                        return errors.Wrap(err, "")
                }
        }
        if _, err := conn.Do("set", key, value, "ex", int64(expiration/time.Second)); err != nil {
                return errors.Wrap(err, "")
        }
        return nil
}

func Get(ctx context.Context, key string) (interface{}, error) {
        conn, err := getConn(ctx)
        if err != nil {
                return nil, err
        }
        defer conn.Close()

        reply, err := conn.Do("get", key)
        if err != nil {
                return nil, errors.Wrap(err, "redis error.")
        }
        return reply, nil
}

func GetStr(ctx context.Context, key string) (string, error) {
        reply, err := Get(ctx, key)
        if err != nil || reply == nil {
                return "", err
        }
        return redis.String(reply, nil)
}
