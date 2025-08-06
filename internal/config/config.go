package config

import (
	"log"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	Env         string 	 	  `mapstructure:"env"`
	DB 			DBConfig 	  `mapstructure:"database" validate:"required"`
	AppSecret   string   	  `mapstructure:"app_secret"`
	HTTPServer  	     	  `mapstructure:"http_server"`
	SSOServer		     	  `mapstructure:"sso_server"`
	Kafka 		KafkaProducer `mapstructure:"kafka_producer"`
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
	SSLKeyLocation 		 		 	   string `mapstructure:"ssl_key_location" validate:"required"`
	SSLCertificateLocation 		 	   string `mapstructure:"ssl_certificate_location" validate:"required"`
	SSLCaLocation		 		 	   string `mapstructure:"ssl_ca_location" validate:"required"`
	SSLEndpointIdentificationAlgorithm string `mapstructure:"ssl_endpoint_identification_algorithm"`
}

func (k KafkaProducer) Get(key string) (string, bool) {
	switch key {
	case "bootstrap.servers":
		return k.BootstrapServers, true
	case "sasl.username":
		return k.SaslUsername, true
	case "sasl.password":
		return k.SaslPassword, true
	case "ssl.key.location":
		return k.SSLKeyLocation, true
	case "ssl.certificate.location":
		return k.SSLCertificateLocation, true
	case "ssl.ca.location":
		return k.SSLCaLocation, true
	case "ssl.endpoint.identification.algorithm":
		return k.SSLEndpointIdentificationAlgorithm, true
	default:
		return "", false
	}
}


func MustLoad() *Config {
	err := godotenv.Load("./config/.env")
	if err != nil {
	 	log.Println(".env file not found or failed to load, skipping: " + err.Error())
	}
	
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
		"database.url":                            "DATABASE_URL",
		"kafka_producer.bootstrap_servers":        "KAFKA_BOOTSTRAP_SERVERS",
		"kafka_producer.sasl_username":            "KAFKA_SASL_USERNAME",
		"kafka_producer.sasl_password":            "KAFKA_SASL_PASSWORD",
		"kafka_producer.ssl_key_location":         "KAFKA_SSL_KEY_LOCATION",
		"kafka_producer.ssl_certificate_location": "KAFKA_SSL_CERTIFICATE_LOCATION",
		"kafka_producer.ssl_ca_location":          "KAFKA_SSL_CA_LOCATION",
	}

	for key, envVar := range envBindings {
		if val := os.Getenv(envVar); val != "" {
			v.Set(key, val)
		}
	}

	var cfg Config

	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("error unmarshaling config: %s", err)
	}

	if err := validator.New().Struct(cfg); err != nil {
		log.Fatalf("error validating config: %s", err)
	}

	return &cfg
}

