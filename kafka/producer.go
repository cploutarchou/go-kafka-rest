package kafka

import (
	"log"
	"sync"

	"github.com/Shopify/sarama"
)

type ProducerFactory func(brokers []string, config *sarama.Config) (sarama.SyncProducer, error)

var (
	instance *Producer
	once     sync.Once
)

type Producer struct {
	producer sarama.SyncProducer
	mutex    *sync.Mutex
}

func NewProducer(brokers []string, conf *sarama.Config, factory ProducerFactory) (*Producer, error) {
	once.Do(func() {
		var config *sarama.Config
		if conf == nil {
			config = sarama.NewConfig()
		} else {
			config = conf
		}
		if config.Producer.RequiredAcks == sarama.NoResponse {
			config.Producer.RequiredAcks = sarama.WaitForAll
		}
		if config.Producer.Retry.Max < 10 {
			config.Producer.Retry.Max = 10
		}

		config.Producer.Return.Successes = true
		config.Producer.Compression = sarama.CompressionSnappy

		producer, err := factory(brokers, config)
		if err != nil {
			log.Fatal("Failed to start producer: ", err)
		}

		instance = &Producer{
			producer: producer,
			mutex:    &sync.Mutex{},
		}
	})

	return instance, nil
}

func (p *Producer) SendMessage(topic string, key string, value string) (partition int32, offset int64, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(value),
	}
	if key != "" {
		msg.Key = sarama.StringEncoder(key)
	}
	partition, offset, err = p.producer.SendMessage(msg)
	return
}
