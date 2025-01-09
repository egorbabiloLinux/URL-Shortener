package config

import (
	"log"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Env         string `mapstructure:"env"`
	StoragePath string `mapstructure:"storage_path"`
	HTTPServer  `mapstructure:"http_server"`
}

type HTTPServer struct {
	Address     string        `mapstructure:"address"`
	Timeout     time.Duration `mapstructure:"timeout"`
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`
	User 		string		  `mapstructure:"user" validate:"required"`
	Password 	string 		  `mapstructure:"password" validate:"required" env:"HTTP_SERVER_PASSWORD"`
}

func MustLoad() *Config {
	v := viper.New()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH enviroment variable is not set")
	}

	v.SetConfigFile(configPath)

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("error reading config file: %s", err)
	}

	v.AutomaticEnv()

	v.SetDefault("env", "local")
	v.SetDefault("http_server.address", "localhost:8082")
	v.SetDefault("http_server.timeout", "4s")
	v.SetDefault("http_server.idle_timeout", "30s")

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("error unmarshalling config: %s", err)
	}

	if err := validator.New().Struct(cfg); err != nil {
		log.Fatalf("error validating config: %s", err)
	}

	return &cfg
}
