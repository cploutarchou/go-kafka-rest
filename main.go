package main

import (
	"fmt"
	"github.com/cploutarchou/go-kafka-rest/controllers"
	"github.com/cploutarchou/go-kafka-rest/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cploutarchou/go-kafka-rest/initializers"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var controller *controllers.Controller

func setupApp() (*fiber.App, error) {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		return nil, fmt.Errorf("failed to load environment variables! \n%s", err.Error())
	}
	initializers.ConnectDB(config)
	brokers := strings.Split(config.KafkaBrokers, ",")
	controller = controllers.NewController(initializers.GetDB(), brokers, 7)

	if config.KafkaBrokers == "" {
		return nil, fmt.Errorf("failed to load environment variables! \n%s", "KAFKA_BROKER is not set")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to Kafka! \n%s", err.Error())
	}

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
		log.Printf("execution time: %v", time.Since(start))
		return nil
	})

	return app, nil
}
func setupMicro(controller *controllers.Controller) *fiber.App {
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
		router.Get("/refresh", middleware.DeserializeUser, controller.User.RefreshToken)
		router.Post("/receive-message", controller.User.ReceiveMessage)
	})

	micro.Get("/users/me", middleware.DeserializeUser, controller.User.GetMe)
	micro.All("*", func(c *fiber.Ctx) error {
		path := c.Path()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "fail",
			"message": fmt.Sprintf("path: %v does not exist on this server", path),
		})
	})

	return micro
}

func main() {
	app, err := setupApp()
	if err != nil {
		log.Fatalf("failed to setup app: %v", err)
		return
	}

	micro := setupMicro(controller)

	app.Mount("/api", micro)

	// Getting port from env variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8045" // Default port if not specified
	}

	go func() {
		if err := app.Listen(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Print("shutdown Server ...")

	if err := app.Shutdown(); err != nil {
		log.Fatal("server Shutdown: ", err)
	}

	log.Print("server exiting")
}
