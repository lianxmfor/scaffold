package db

import (
        "context"
        "fmt"
        "time"

        "github.com/gomodule/redigo/redis"
        "github.com/pkg/errors"
)

type RedisConfig struct {
        Addr        string
        Password    string
        DB          int
        MaxActive   int
        MaxIdle     int
        DialTimeout Duration
}

type Duration struct {
        time.Duration
}

// UnmarshalText 将字符串形式的时长信息转换为Duration类型
func (d *Duration) UnmarshalText(text []byte) (err error) {
        d.Duration, err = time.ParseDuration(string(text))
        return
}

// D 从Duration struct中取出time.Duration类型的值
func (d *Duration) D() time.Duration {
        return d.Duration
}

var Redis *redis.Pool

func InitRedis(conf RedisConfig) error {
        var options = []redis.DialOption{
                redis.DialDatabase(conf.DB),
                redis.DialPassword(conf.Password),
        }
        if conf.DialTimeout.D()> 0 {
                options = append(options, redis.DialConnectTimeout(conf.DialTimeout.D()))
        }

        pool := &redis.Pool{
                Dial: func() (conn redis.Conn, e error) {
                        c, err := redis.Dial("tcp", conf.Addr, options...)
                        if err != nil {
                                return nil, err
                        }
                        return c, nil
                },
                TestOnBorrow: func(c redis.Conn, t time.Time) error {
                        _, err := c.Do("PING")
                        return err
                },
                MaxIdle:   conf.MaxIdle,
                MaxActive: conf.MaxActive,
        }

        if pool.MaxActive > 0 {
                pool.Wait = true
        }
        Redis = pool
        return nil
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
