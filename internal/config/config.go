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
	AppSecret string   `mapstructure:"app_secret"`
	HTTPServer  	   `mapstructure:"http_server"`
	SSOServer		   `mapstructure:"sso_server"`
}

type SSOServer struct {
	SSOAddr 	string 		  `mapstructure:"grpc_addr"`
	SSOTimeout time.Duration  `mapstructure:"grpc_timeout"`
	SSORetries 	int 		  `mapstructure:"retries"`
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
	//v.SetDefault("app_secret", "default_app_secret") - не стоит добавлять в целях безопасности
	v.SetDefault("http_server.address", "localhost:8082")
	v.SetDefault("http_server.timeout", "4s")
	v.SetDefault("http_server.idle_timeout", "30s")

	v.SetDefault("sso_server.grpc_addr", "localhost:44044")
	v.SetDefault("sso_server.grpc_timeout", "5s")
	v.SetDefault("sso_server.retries", 3)

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("error unmarshalling config: %s", err)
	}

	if err := validator.New().Struct(cfg); err != nil {
		log.Fatalf("error validating config: %s", err)
	}

	return &cfg
}
