package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/cploutarchou/go-kafka-rest/controllers"
	"github.com/cploutarchou/go-kafka-rest/initializers"
	middlewares "github.com/cploutarchou/go-kafka-rest/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

var (
	controller *controllers.Controller
	app        *fiber.App
	middleware *middlewares.Middleware
)

func setupApp() (*fiber.App, error) {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %s", err.Error())
	}
	initializers.ConnectDB(config)
	db := initializers.GetDB()
	middleware = middlewares.NewMiddleware(config, db)
	brokers := strings.Split(config.KafkaBrokers, ",")
	controller = controllers.NewController(db, brokers, int32(config.KafkaNumOfPartitions))
	if config.KafkaBrokers == "" {
		return nil, fmt.Errorf("KAFKA_BROKER environment variable is not set")
	}
	app := fiber.New()
	app.Use(logger.New())
	allowedOrigins := strings.Join(config.CorsAllowedOrigins, ",")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
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

func setupRoutes(controller *controllers.Controller) *fiber.App {
	app := fiber.New()
	app.Get("/healthchecker", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "JWT Authentication with Golang, Fiber, and GORM",
		})
	})

	app.Route("/auth", func(router fiber.Router) {
		router.Post("/register", controller.User.SignUpUser)
		router.Post("/login", controller.User.SignInUser)
		router.Get("/logout", middleware.DeserializeUser, controller.User.LogoutUser)
		router.Get("/refresh", middleware.DeserializeUser, controller.User.RefreshToken)
	})

	app.Route("/kafka", func(router fiber.Router) {
		router.Post("/send-message", middleware.DeserializeUser, controller.User.SendMessage)
	})

	app.Get("/users/me", middleware.DeserializeUser, controller.User.GetMe)

	app.All("*", func(c *fiber.Ctx) error {
		path := c.Path()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "fail",
			"message": fmt.Sprintf("path: %v does not exist on this server", path),
		})
	})

	return app
}

func main() {
	var err error
	app, err = setupApp()
	if err != nil {
		log.Fatalf("failed to setup app: %v", err)
		return
	}

	routes := setupRoutes(controller)

	app.Mount("/api", routes)

	// Get the port from the environment variable or use the default port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8045"
	}

	go func() {
		if err := app.Listen(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server listen: %s", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Print("shutting down server...")

	if err := app.Shutdown(); err != nil {
		log.Fatal("server shutdown:", err)
	}

	log.Print("server exited")
}
