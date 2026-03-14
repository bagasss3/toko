package cache

import (
    "fmt"
    "time"

    "github.com/gomodule/redigo/redis"
    "github.com/kumparan/cacher"
)

type Config struct {
    Host       string
    Password   string
    DB         int
    DefaultTTL time.Duration
    MaxIdle    int
    MaxActive  int
}

func newPool(cfg Config) *redis.Pool {
    return &redis.Pool{
        MaxIdle:     cfg.MaxIdle,
        MaxActive:   cfg.MaxActive,
        IdleTimeout: 240 * time.Second,
        Dial: func() (redis.Conn, error) {
            url := fmt.Sprintf("redis://%s/%d", cfg.Host, cfg.DB)
            c, err := redis.DialURL(url)
            if err != nil {
                return nil, err
            }
            if cfg.Password != "" {
                if _, err := c.Do("AUTH", cfg.Password); err != nil {
                    c.Close()
                    return nil, err
                }
            }
            return c, nil
        },
    }
}

func NewKeeper(cfg Config) cacher.Keeper {
    pool := newPool(cfg)

    k := cacher.NewKeeper()
    k.SetConnectionPool(pool)
    k.SetLockConnectionPool(pool) 
    k.SetDefaultTTL(cfg.DefaultTTL)

    return k
}