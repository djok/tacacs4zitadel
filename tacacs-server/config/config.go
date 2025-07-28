package config

import (
	"github.com/spf13/viper"
	"github.com/sirupsen/logrus"
)

type Config struct {
	TACACSListenAddress string `mapstructure:"tacacs_listen_address"`
	HTTPListenAddress   string `mapstructure:"http_listen_address"`
	TACACSSecret        string `mapstructure:"tacacs_secret"`
	LogLevel            string `mapstructure:"log_level"`
	
	// Zitadel Configuration
	ZitadelURL          string `mapstructure:"zitadel_url"`
	ZitadelProjectID    string `mapstructure:"zitadel_project_id"`
	ZitadelClientID     string `mapstructure:"zitadel_client_id"`
	ZitadelClientSecret string `mapstructure:"zitadel_client_secret"`
	
	DBHost     string `mapstructure:"db_host"`
	DBPort     string `mapstructure:"db_port"`
	DBName     string `mapstructure:"db_name"`
	DBUser     string `mapstructure:"db_user"`
	DBPassword string `mapstructure:"db_password"`
	
	SessionTimeout        int `mapstructure:"session_timeout"`
	TokenCacheTimeout     int `mapstructure:"token_cache_timeout"`
	MaxConcurrentSessions int `mapstructure:"max_concurrent_sessions"`
}

func Load() *Config {
	viper.SetDefault("tacacs_listen_address", "0.0.0.0:49")
	viper.SetDefault("http_listen_address", "0.0.0.0:8090")
	viper.SetDefault("tacacs_secret", "testing123")
	viper.SetDefault("log_level", "info")
	
	// Zitadel defaults
	viper.SetDefault("zitadel_url", "http://localhost:8080")
	viper.SetDefault("zitadel_project_id", "")
	viper.SetDefault("zitadel_client_id", "")
	viper.SetDefault("zitadel_client_secret", "")
	
	viper.SetDefault("db_host", "localhost")
	viper.SetDefault("db_port", "5432")
	viper.SetDefault("db_name", "zitadel")
	viper.SetDefault("db_user", "zitadel")
	viper.SetDefault("db_password", "zitadel")
	
	viper.SetDefault("session_timeout", 3600)
	viper.SetDefault("token_cache_timeout", 300)
	viper.SetDefault("max_concurrent_sessions", 1000)

	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		logrus.WithError(err).Fatal("Failed to unmarshal configuration")
	}

	return &config
}