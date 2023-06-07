package controllers

import (
	"github.com/Shopify/sarama"
	"github.com/cploutarchou/go-kafka-rest/kafka"
	"github.com/cploutarchou/go-kafka-rest/models"
	"github.com/cploutarchou/go-kafka-rest/types"
	"gorm.io/gorm"
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
	Auth            AuthController
	User            UserController
	workerPoolSize  int
	workPool        chan struct{}
	wg              *sync.WaitGroup
	mutex           *sync.Mutex
	messageQueue    []types.MessagePayload
	producer        kafka.Producer
	SaramaConfig    *sarama.Config
	DB              *gorm.DB
	totalPartitions int32
}

func NewController(db *gorm.DB, brokers_ []string, partitions int32) *Controller {

	con := &Controller{
		User:            NewUserController(db),
		workerPoolSize:  workerPoolSize,
		workPool:        workerPool,
		wg:              &wg,
		mutex:           &mutex,
		messageQueue:    messageQueue,
		totalPartitions: partitions,
	}
	brokers = brokers_
	con.Initialize()

	return con
}

func NewUserController(db *gorm.DB) UserController {
	userCon := UserController{
		DB:    db,
		Model: models.User{},
	}
	userCon.Model.SetDB(db)
	return userCon

}

func (c *Controller) Initialize() {
	var err error
	if c.SaramaConfig == nil {
		c.SaramaConfig = sarama.NewConfig()
		producer, err = kafka.NewProducer(brokers, c.SaramaConfig, kafka.TheProducerFactory)
		if err != nil {
			panic(err)
		}
	} else {
		producer, err = kafka.NewProducer(brokers, c.SaramaConfig, kafka.TheProducerFactory)
		if err != nil {
			panic(err)
		}
	}

}

func (c *Controller) SetBrokers(b []string) {
	brokers = b
}

func (c *Controller) SetSaramaConfig(config *sarama.Config) {
	c.SaramaConfig = config
}

func (c *Controller) SetWorkerPoolSize(size int) {
	c.workerPoolSize = size
}
