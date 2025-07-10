package kafka

import "github.com/confluentinc/confluent-kafka-go/kafka"

func NewProducer() (*kafka.Producer, error) {
	return kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers" : "kafka1:9092,kafka2:9093, kafka3:9094",
		"security.protocol" : "SASL_SSL",
		"sasl.mechanisms" : "PLAIN",
		"sasl.username" : "va",
		"sasl.password" : "222222",
		"ssl.keystore.location" : "./keystore/kafka.client.keystore.jks", //TODO переделать в pem
		"ssl.keystore.password" : "supersecret",
		"ssl.truststore.location" : "./keystore/kafka.client.truststore.jks", //TODO переделать в pem
		"ssl.truststore.password" : "supersecret",
		"ssl.endpoint.identification.algorithm" : "", //TODO проверить 
	})
}
