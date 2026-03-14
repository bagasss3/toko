package console

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	pkgconfig "github.com/bagasss3/toko/packages/config"
	"github.com/bagasss3/toko/packages/logger"

	"github.com/bagasss3/toko/services/gateway/internal/client"
	"github.com/bagasss3/toko/services/gateway/internal/config"
	"github.com/bagasss3/toko/services/gateway/internal/delivery"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start HTTP gateway server",
	Run:   runServer,
}

func runServer(cmd *cobra.Command, args []string) {
	log := logger.NewEntry("gateway")

	pkgconfig.Init()
	cfg := config.Load()

	authClient, err := client.NewAuthClient(cfg.AuthServiceAddr)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to auth-service")
	}
	defer authClient.Close()

	srv := delivery.NewServer(authClient)

	go func() {
		log.Infof("gateway listening on :%s", cfg.Port)
		if err := srv.Start(fmt.Sprintf(":%s", cfg.Port)); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("gateway server failed")
		}
	}()

	// block until OS signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down gateway...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Warn("gateway forced to shutdown")
	}

	log.Info("gateway exited")
}
