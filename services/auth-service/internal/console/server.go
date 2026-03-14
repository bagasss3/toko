package console

import (
	"fmt"
	"time"

	"github.com/bagasss3/toko/packages/cache"
	pkgconfig "github.com/bagasss3/toko/packages/config"
	"github.com/bagasss3/toko/packages/database"
	"github.com/bagasss3/toko/packages/logger"
	"github.com/spf13/cobra"

	"github.com/bagasss3/toko/services/auth-service/internal/config"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start HTTP server",
	Run:   runServer,
}

func runServer(cmd *cobra.Command, args []string) {
	log := logger.NewEntry("auth-service")

	pkgconfig.Init()
	cfg := config.Load()

	db, err := database.Init(database.Config{
		DSN:             fmt.Sprintf("postgresql://%s:%s@%s/%s", cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBName),
		MaxIdleConns:    3,
		MaxOpenConns:    15,
		ConnMaxLifetime: 30 * time.Minute,
		ConnMaxIdleTime: 10 * time.Minute,
		PingInterval:    10 * time.Second,
		RetryAttempts:   5,
	})
	if err != nil {
		log.WithError(err).Fatal("failed to init database")
	}
	defer db.Close()

	_ = cache.NewKeeper(cache.Config{
		Host:       cfg.RedisHost,
		Password:   cfg.RedisPassword,
		DB:         cfg.RedisDB,
		DefaultTTL: 15 * time.Minute,
		MaxIdle:    100,
		MaxActive:  10000,
	})

	log.Infof("starting on port %s", cfg.Port)
}
