package controllers

import (
	"github.com/Shopify/sarama"
	"github.com/cploutarchou/go-kafka-rest/kafka"
	"github.com/cploutarchou/go-kafka-rest/types"
	"sync"
)

var (
	workerPoolSize = 100 // Number of workers in the pool
	workerPool     = make(chan struct{}, workerPoolSize)
	wg             sync.WaitGroup         // WaitGroup to wait for workers to finish
	mutex          sync.Mutex             // Mutex to protect shared resources
	messageQueue   []types.MessagePayload // Shared message queue
	producer       *kafka.Producer        // Kafka producer
	brokers        []string               // Kafka brokers
)

type Controller struct {
	Auth           AuthController
	User           UserController
	workerPoolSize int
	workPool       chan struct{}
	wg             sync.WaitGroup
	mutex          sync.Mutex
	messageQueue   []types.MessagePayload
	producer       kafka.Producer
	SaramaConfig   *sarama.Config
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) Init() {
	var err error
	if c.SaramaConfig == nil {
		c.SaramaConfig = sarama.NewConfig()
		producer, err = kafka.NewProducer(brokers, c.SaramaConfig, myProducerFactory)
		if err != nil {
			panic(err)
		}
	} else {
		producer, err = kafka.NewProducer(brokers, c.SaramaConfig, myProducerFactory)
		if err != nil {
			panic(err)
		}
	}

}

func (c *Controller) SetBrokers(b []string) {
	brokers = b
}

func myProducerFactory(brokers []string, config *sarama.Config) (sarama.SyncProducer, sarama.AsyncProducer, error) {
	syncProducer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, nil, err
	}

	asyncProducer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		err := syncProducer.Close()
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, err
	}

	return syncProducer, asyncProducer, nil
}
