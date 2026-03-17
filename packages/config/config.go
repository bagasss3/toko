package config

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func Init(paths ...string) {
	if len(paths) == 0 {
		paths = []string{"."}
	}

	for _, path := range paths {
		viper.AddConfigPath(path)
	}

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	if err := viper.ReadInConfig(); err != nil {
		logrus.Warnf("config file not found: %v", err)
	}

	if viper.ConfigFileUsed() != "" {
		logrus.Infof("using config file: %s", viper.ConfigFileUsed())
	}
}

func GetString(key, fallback string) string {
	if !viper.IsSet(key) {
		return fallback
	}
	return viper.GetString(key)
}

func GetInt(key string, fallback int) int {
	if !viper.IsSet(key) {
		return fallback
	}
	return viper.GetInt(key)
}

func GetBool(key string, fallback bool) bool {
	if !viper.IsSet(key) {
		return fallback
	}
	return viper.GetBool(key)
}

func GetDuration(key string, fallback time.Duration) time.Duration {
	raw := viper.GetString(key)
	if raw == "" {
		return fallback
	}
	d, err := time.ParseDuration(raw)
	if err != nil {
		return fallback
	}
	return d
}
