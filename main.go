package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
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

var producer *kafka.Producer

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
	var messagePayload map[string]interface{}
	err := c.BodyParser(&messagePayload)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Invalid request payload",
		})
	}

	// Process the message or perform any validations here
	// check if topic exists in messagePayload and if data is not empty
	if _, ok := messagePayload["topic"]; !ok {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Topic is missing",
		})
	}

	if _, ok := messagePayload["data"]; !ok {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Data is missing",
		})
	}

	// check if key exists in messagePayload
	if _, ok := messagePayload["key"]; !ok {
		messagePayload["key"] = ""
	}
	type Message struct {
		Topic string `json:"topic"`
		Data  string `json:"data"`
		Key   string `json:"key"`
	}

	msg := &Message{
		Topic: messagePayload["topic"].(string),
		Data:  messagePayload["data"].(string),
	}

	producer.SendMessageAsync(msg.Topic, msg.Data, msg.Key)

	log.Printf("Kafka message produced! Topic: %s\n", msg.Topic)

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Message received and Kafka message produced",
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
