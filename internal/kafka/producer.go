package kafka

import (
	"URL-Shortener/internal/event"
	"URL-Shortener/internal/lib/logger/sl"
	"fmt"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type ProducerConfig interface {
	Get(key string) (string, bool)
}

type KafkaProducer struct {
	log *slog.Logger
	*kafka.Producer
}


func NewProducer(cfg ProducerConfig, log *slog.Logger) (*KafkaProducer, error) {
	get := func(key string) string {
		val, _ := cfg.Get(key)
		return val
	}

	p, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers" : 		get("bootstrap.servers"),
		"security.protocol" : 		"SASL_SSL",
		
		"sasl.mechanisms" : 		"PLAIN",
		"sasl.username" : 			get("sasl.username"),
		"sasl.password" : 			get("sasl.password"),
		
		"ssl.key.location": 		get("ssl.key.location"),
		"ssl.certificate.location": get("ssl.certificate.location"),
		"ssl.ca.location": 			get("ssl.ca.location"),
	})
	if err != nil {
		return nil, err 
	}

	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
				case *kafka.Message:
					if ev.TopicPartition.Error != nil {
						log.Error("Failed to deliver message", 
						slog.String("partition", *ev.TopicPartition.Topic))
					} else {
						log.Info(fmt.Sprintf("Produced event to topic %s: key = %s, value = %s",
						slog.String("topic", *ev.TopicPartition.Topic), 
						slog.String("key", string(ev.Key)), 
						slog.String("key", string(ev.Value))))
					}
			}
		}
	}()

	return &KafkaProducer{
		log: log,
		Producer: p,
	}, nil
}

func (p *KafkaProducer) ProduceEvent(ev event.Event, topic string) error {
	const op = "event.ProduceEvent"

	value, err := ev.Value()
	if err != nil {
		p.log.Error("failed to get event value", 
		slog.String("op", op), 
		sl.Err(err))
	}

	return p.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key: ev.Key(),
		Value: value,
	}, nil)
}

func (p *KafkaProducer) CloseProducer() {
	const op = "kafka.KafkaProduce.Close"

	remaining := p.Flush(15 * 1000)
	if remaining > 0 {
		p.log.Warn("not all messages was delivered", 
		slog.String("op", op),
		slog.Int("remaining", remaining))
	}

	p.Close()
	p.log.Info("kafka producer closed")
}
