package config

import (
	"time"

	"github.com/bagasss3/toko/packages/config"
)

type Config struct {
	Port                string
	DBHost              string
	DBUser              string
	DBPassword          string
	DBName              string
	RedisHost           string
	RedisPassword       string
	RedisDB             int
	RedisExpired        time.Duration
	JWTKey              string
	PasetoKey           string
	AccessTokenDuration time.Duration
}

func Load() Config {
	config.Init()

	return Config{
		Port:                config.GetString("port", "8080"),
		DBHost:              config.GetString("database.host", "localhost"),
		DBUser:              config.GetString("database.username", "postgres"),
		DBPassword:          config.GetString("database.password", "postgres"),
		DBName:              config.GetString("database.database", "toko_auth"),
		RedisHost:           config.GetString("redis.host", "localhost:6379"),
		RedisPassword:       config.GetString("redis.password", ""),
		RedisDB:             config.GetInt("redis.db", 0),
		RedisExpired:        config.GetDuration("redis.exp", 24*time.Hour),
		JWTKey:              config.GetString("jwt.key", ""),
		PasetoKey:           config.GetString("paseto.key", "12345678901234567890123456789012"),
		AccessTokenDuration: config.GetDuration("paseto.access_token_duration", 15*time.Minute),
	}
}
