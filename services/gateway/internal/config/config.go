package config

import (
	"time"

	pkgconfig "github.com/bagasss3/toko/packages/config"
)

type Config struct {
	Port            string
	AuthServiceAddr string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

func Load() Config {
	return Config{
		Port:            pkgconfig.GetString("port", "8080"),
		AuthServiceAddr: pkgconfig.GetString("services.auth_addr", "localhost:9001"),
		ReadTimeout:     pkgconfig.GetDuration("server.read_timeout", 10*time.Second),
		WriteTimeout:    pkgconfig.GetDuration("server.write_timeout", 10*time.Second),
	}
}