package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"backend/pkg/log"
	"github.com/spf13/viper"
)

type Config struct {
	PGName       string
	PGUser       string
	PGPassword   string
	PGHost       string
	PGPort       int
	MaxPool      int32
	PGTimeout    time.Duration
	ConnAttempts int
	ServiceHost  string
	ServicePort  string
	IsTest       bool
}

const (
	PGName       = "PG_NAME"
	PGUser       = "PG_USER"
	PGPassword   = "PG_PASSWORD"
	PGHost       = "PG_HOST"
	PGPort       = "PG_PORT"
	PGTimeout    = "PG_TIMEOUT"
	PGMaxPool    = "PG_MAX_POOL"
	ConnAttempts = "CONN_ATTEMPTS"
	ServiceHost  = "SERVICE_HOST"
	ServicePort  = "SERVICE_PORT"
)

const (
	_defaultServiceHost = "localhost"
	_defaultServicePort = "8080"
)

func InitConfig() *Config {
	envPath, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("err getting work dir: %v", err.Error()))
	}

	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(envPath)

	viper.AutomaticEnv()

	viper.SetDefault(ServiceHost, _defaultServiceHost)
	viper.SetDefault(ServicePort, _defaultServicePort)

	err = viper.ReadInConfig()
	if err != nil {
		if errors.As(err, &viper.ConfigFileNotFoundError{}) {
			log.Log.Info(fmt.Sprintf("config file not found: %v", envPath))
		} else {
			panic(fmt.Sprintf("err reading config: %v", err.Error()))
		}
	}

	return &Config{
		PGName:       viper.GetString(PGName),
		PGUser:       viper.GetString(PGUser),
		PGPassword:   viper.GetString(PGPassword),
		PGHost:       viper.GetString(PGHost),
		PGPort:       viper.GetInt(PGPort),
		MaxPool:      viper.GetInt32(PGMaxPool),
		PGTimeout:    viper.GetDuration(PGTimeout),
		ConnAttempts: viper.GetInt(ConnAttempts),
		ServiceHost:  viper.GetString(ServiceHost),
		ServicePort:  viper.GetString(ServicePort),
	}
}
