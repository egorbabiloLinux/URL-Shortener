package config

import (
	"log"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type Config struct {
	Env         string 	 `mapstructure:"env"`
	DB 			DBConfig `mapstructure:"database" validate:"required"`
	AppSecret   string   `mapstructure:"app_secret"`
	HTTPServer  	     `mapstructure:"http_server"`
	SSOServer		     `mapstructure:"sso_server"`
	KafkaProducer 		 `mapstructure:"kafka_producer"`
}

type DBConfig struct {
	URL string `mapstructure:"url" validate:"required"`
}

type SSOServer struct {
	SSOAddr 	string 		  `mapstructure:"grpc_addr"`
	SSOTimeout  time.Duration `mapstructure:"grpc_timeout"`
	SSORetries 	int 		  `mapstructure:"retries"`
}

type HTTPServer struct {
	Address     string        `mapstructure:"address"`
	Timeout     time.Duration `mapstructure:"timeout"`
	IdleTimeout time.Duration `mapstructure:"idle_timeout"`
	User 		string		  `mapstructure:"user" validate:"required"`
	Password 	string 		  `mapstructure:"password" validate:"required"`
}

type KafkaProducer struct {
	BootstrapServers 				   string `mapstructure:"bootstrap_servers" validate:"required"`
	SaslUsername 					   string `mapstructure:"sasl_username" validate:"required"`
	SaslPassword 					   string `mapstructure:"sasl_password" validate:"required"`
	SSLKeystoreLocation 		       string `mapstructure:"ssl_keystore_location" validate:"required"`
	SSLKeystorePassword 			   string `mapstructure:"ssl_keystore_password" validate:"required"`
	SSLTruststoreLocation 		       string `mapstructure:"ssl_truststore_location" validate:"required"`
	SSLTruststorePassword 			   string `mapstructure:"ssl_truststore_password" validate:"required"`
	SSLEndpointIdentificationAlgorithm string `mapstructure:"ssl_endpoint_identification_algorithm" validate:"required"`
}

func MustLoad() *Config {
	v := viper.New()

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exists: " + configPath)
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

	envBindings := map[string]string{
		"database.url":                          "DATABASE_URL",
		"kafka_producer.bootstrap_servers":      "KAFKA_BOOTSTRAP_SERVERS",
		"kafka_producer.sasl_username":          "KAFKA_SASL_USERNAME",
		"kafka_producer.sasl_password":          "KAFKA_SASL_PASSWORD",
		"kafka_producer.ssl_keystore_location":  "KAFKA_SSL_KEYSTORE_LOCATION",
		"kafka_producer.ssl_keystore_password":  "KAFKA_SSL_KEYSTORE_PASSWORD",
		"kafka_producer.ssl_truststore_location":"KAFKA_SSL_TRUSTSTORE_LOCATION",
		"kafka_producer.ssl_truststore_password":"KAFKA_SSL_TRUSTSTORE_PASSWORD",
	}

	for key, envVar := range envBindings {
		if val := os.Getenv(envVar); val != "" {
			v.Set(key, val)
		}
	}

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("error unmarshalling config: %s", err)
	}

	if err := validator.New().Struct(cfg); err != nil {
		log.Fatalf("error validating config: %s", err)
	}

	return &cfg
}

