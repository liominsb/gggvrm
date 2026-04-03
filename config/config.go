package config

import (
	"log"

	"github.com/spf13/viper"
)

// Config 配置项
type Config struct {
	App struct {
		Name string
		Port string
	}
	Database struct {
		Dsn          string
		MaxIdleConns int
		MaxOpenConns int
		Addr         string
		Password     string
	}
}

var Appconf *Config

func InitConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	Appconf = &Config{}

	if err := viper.Unmarshal(Appconf); err != nil {
		log.Fatalf("Failed to unmarshal config file: %v", err)
	}

	initDB()
	initRedis()
}
