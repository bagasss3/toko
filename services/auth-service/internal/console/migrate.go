package console

import (
	"fmt"
	"time"

	"github.com/bagasss3/toko/packages/database"
	"github.com/pressly/goose/v3"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/bagasss3/toko/packages/config"
	svcconfig "github.com/bagasss3/toko/services/auth-service/internal/config"
)

var migrateCmd = &cobra.Command{
	Use:     "migrate",
	Aliases: []string{"m"},
	Short:   "run database migrations",
	Long:    "Run database migrations using goose",
	Run:     runMigrate,
}

func init() {
	migrateCmd.PersistentFlags().String("direction", "up", "migration direction: up or down")
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) {
	direction, _ := cmd.Flags().GetString("direction")

	log := logrus.New()
	log.SetFormatter(&logrus.JSONFormatter{})

	config.Init()
	cfg := svcconfig.Load()

	// Set Goose dialect to PostgreSQL
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set Goose dialect: %v", err)
	}

	goose.SetTableName("schema_migrations")

	// Initialize database using the shared package
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
		log.Fatalf("Failed to init database: %v", err)
	}
	defer db.Close()

	// Get underlying *sql.DB from GORM
	sqlDB, err := db.Conn.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}

	migrationDir := "./db/migration"

	switch direction {
	case "up":
		err = goose.Up(sqlDB, migrationDir)
	case "down":
		err = goose.Down(sqlDB, migrationDir)
	default:
		log.Fatalf("Unknown migration direction: %s (use 'up' or 'down')", direction)
	}

	if err != nil {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	log.Infof("Migrations applied successfully: %s", direction)
}
