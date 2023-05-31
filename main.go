package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
	
	"github.com/Shopify/sarama"
	"github.com/cploutarchou/go-kafka-rest/kafka"
	
	"github.com/cploutarchou/go-kafka-rest/controllers"
	"github.com/cploutarchou/go-kafka-rest/initializers"
	"github.com/cploutarchou/go-kafka-rest/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type MessagePayload struct {
	Topic string `json:"topic"`
	Data  string `json:"data"`
	Key   string `json:"key"`
}

var (
	workerPoolSize = 100 // Number of workers in the pool
	workerPool     = make(chan struct{}, workerPoolSize)
	wg             sync.WaitGroup   // WaitGroup to wait for workers to finish
	mutex          sync.Mutex       // Mutex to protect shared resources
	messageQueue   []MessagePayload // Shared message queue
	producer       *kafka.Producer  // Kafka producer 
)

func init() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load environment variables! \n%s", err.Error())
	}
	initializers.ConnectDB(config)
	// check if broker is set in environment variables
	if config.KafkaBrokers == "" {
		log.Fatalf("Failed to load environment variables! \n%s", "KAFKA_BROKER is not set")
	}
	brokers := strings.Split(config.KafkaBrokers, ",")
	producer, err = kafka.NewProducer(brokers, nil, myProducerFactory)
	
	if err != nil {
		log.Fatalf("Failed to connect to Kafka! \n%s", err.Error())
	}
}

func setupApp() *fiber.App {
	app := fiber.New()
	
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowMethods:     "GET, POST",
		AllowCredentials: true,
	}))
	
	app.Use(func(c *fiber.Ctx) error {
		start := time.Now()
		if err := c.Next(); err != nil {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"status":  "fail",
				"message": err.Error(),
			})
		}
		log.Printf("[%s] %s", c.Method(), c.Path())
		log.Printf("Execution time: %v", time.Since(start))
		return nil
	})
	
	return app
}

func setupMicro() *fiber.App {
	micro := fiber.New()
	
	micro.Get("/healthchecker", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "JWT Authentication with Golang, Fiber, and GORM",
		})
	})
	
	micro.Route("/auth", func(router fiber.Router) {
		router.Post("/register", controllers.SignUpUser)
		router.Post("/login", controllers.SignInUser)
		router.Get("/logout", middleware.DeserializeUser, controllers.LogoutUser)
		router.Post("/receive-message", receiveMessage)
	})
	
	micro.Get("/users/me", middleware.DeserializeUser, controllers.GetMe)
	micro.All("*", func(c *fiber.Ctx) error {
		path := c.Path()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "fail",
			"message": fmt.Sprintf("Path: %v does not exist on this server", path),
		})
	})
	
	return micro
}

func main() {
	app := setupApp()
	micro := setupMicro()
	
	app.Mount("/api", micro)
	
	log.Fatal(app.Listen(":8045"))
}

func receiveMessage(c *fiber.Ctx) error {
	// Parse JSON payload
	var messagePayload MessagePayload
	err := c.BodyParser(&messagePayload)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Invalid request payload",
		})
	}
	
	// Validate the message payload
	if messagePayload.Topic == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Topic is missing",
		})
	}
	
	if messagePayload.Data == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Data is missing",
		})
	}
	
	// Add the message to the queue
	mutex.Lock()
	messageQueue = append(messageQueue, messagePayload)
	mutex.Unlock()
	
	// Notify a worker to process the message
	workerPool <- struct{}{}
	wg.Add(1)
	
	go func() {
		defer func() {
			// Release the worker and mark the task as done
			<-workerPool
			wg.Done()
		}()
		
		// Process messages from the queue
		for {
			// Acquire a message from the queue
			mutex.Lock()
			if len(messageQueue) == 0 {
				mutex.Unlock()
				break
			}
			message := messageQueue[0]
			messageQueue = messageQueue[1:]
			mutex.Unlock()
			
			// Process the message
			producer.SendMessageAsync(message.Topic, message.Data, message.Key)
			log.Printf("Kafka message produced! Topic: %s\n", message.Topic)
		}
	}()
	
	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Message received and added to the processing queue",
	})
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
