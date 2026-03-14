package console

import (
    "github.com/spf13/cobra"
    log "github.com/sirupsen/logrus"
)

var rootCmd = &cobra.Command{
    Use:   "auth-service",
    Short: "Toko auth service",
}

func Execute() {
    if err := rootCmd.Execute(); err != nil {
        log.Fatal(err)
    }
}

func init() {
    
}