package kafka

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Shopify/sarama"
)

var once sync.Once
var instance *Producer
var initErr error

type ProducerFactory func(brokers []string, conf *sarama.Config) (sarama.SyncProducer, sarama.AsyncProducer, error)

type Producer struct {
	syncProducer  sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
	mutex         *sync.Mutex
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

		syncProducer, asyncProducer, err := factory(brokers, config)
		if err != nil {
			initErr = fmt.Errorf("failed to start producer: %w", err)
			return
		}

		instance = &Producer{
			syncProducer:  syncProducer,
			asyncProducer: asyncProducer,
			mutex:         &sync.Mutex{},
		}

		log.Println("🚀 successfully connected to Kafka producer")
	})

	if initErr != nil {
		return nil, initErr
	}
	if instance == nil {
		return nil, errors.New("producer is not initialized")
	}
	return instance, nil
}

func (p *Producer) SendMessageSync(topic string, key string, value string) (partition int32, offset int64, err error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.StringEncoder(value),
		Timestamp: time.Now().UTC(),
		Offset:    sarama.OffsetNewest,
	}
	if key != "" {
		msg.Key = sarama.StringEncoder(key)
	}
	partition, offset, err = p.syncProducer.SendMessage(msg)
	return
}

func (p *Producer) SendMessageAsync(topic string, key string, value string) {
	// autos-elect partition and offset for message
	msg := &sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.StringEncoder(value),
		Timestamp: time.Now().UTC(),
		Offset:    sarama.OffsetNewest,
	}
	if key != "" {
		msg.Key = sarama.StringEncoder(key)
	}
	p.asyncProducer.Input() <- msg

	go func() {
		select {
		case <-p.asyncProducer.Successes():
			// Message successfully delivered
		case err := <-p.asyncProducer.Errors():
			log.Printf("failed to produce message: %v\n", err.Err)
		}
	}()

}

func (p *Producer) Close() error {
	if err := p.syncProducer.Close(); err != nil {
		log.Fatalln("failed to shut down sync producer cleanly", err)
	}
	if err := p.asyncProducer.Close(); err != nil {
		log.Fatalln("failed to shut down async producer cleanly", err)
	}
	return nil
}
