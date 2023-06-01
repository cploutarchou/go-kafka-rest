package main

import (
	"fmt"
	"github.com/cploutarchou/go-kafka-rest/controllers"
	"log"
	"strings"
	"time"

	"github.com/cploutarchou/go-kafka-rest/initializers"
	"github.com/cploutarchou/go-kafka-rest/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var controller *controllers.Controller

func init() {

	controller = controllers.NewController()
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
	controller.SetBrokers(brokers)
	controller = controllers.NewController()
	controller.Initialize()
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
	// set brokers as fiber app variable

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
		router.Post("/register", controller.User.SignUpUser)
		router.Post("/login", controller.User.SignInUser)
		router.Get("/logout", middleware.DeserializeUser, controller.User.LogoutUser)
		router.Post("/receive-message", controller.User.ReceiveMessage)
	})

	micro.Get("/users/me", middleware.DeserializeUser, controller.User.GetMe)
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
