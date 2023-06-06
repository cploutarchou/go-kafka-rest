package main

import (
	"fmt"
	"github.com/cploutarchou/go-kafka-rest/hub"
	"github.com/gofiber/websocket/v2"
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

// Global variables for controllers, middleware, and fiber app
var (
	controller *controllers.Controller
	app        *fiber.App
	middleware *middlewares.Middleware
	config     *initializers.Config
	brokers    []string
)

// setupApp initializes the fiber app, middleware, and controllers
// It returns a new fiber app and an error if one occurs during setup
func setupApp() (*fiber.App, error) {
	var err error
	// Load the configuration from environment variables
	config, err = initializers.LoadConfig(".")
	if err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %s", err.Error())
	}
	// Connect to the database
	initializers.ConnectDB(config)
	db := initializers.GetDB()

	// Initialize the middleware and controllers
	middleware = middlewares.NewMiddleware(config, db)
	brokers = strings.Split(config.KafkaBrokers, ",")
	controller = controllers.NewController(db, brokers, int32(config.KafkaNumOfPartitions))

	// Check if Kafka brokers are set
	if config.KafkaBrokers == "" {
		return nil, fmt.Errorf("KAFKA_BROKER environment variable is not set")
	}

	// Create a new fiber app
	app := fiber.New()
	// Use the logger middleware for request logging
	app.Use(logger.New())
	// Set the CORS policy
	allowedOrigins := strings.Join(config.CorsAllowedOrigins, ",")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowHeaders:     "Origin, Content-Type, Accept",
		AllowMethods:     "GET, POST",
		AllowCredentials: true,
	}))

	// Custom middleware to log the execution time of each request
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

// setupRoutes sets up all the routes for the fiber app
func setupRoutes(controller *controllers.Controller) *fiber.App {
	app := fiber.New()
	// Health check endpoint
	app.Get("/healthchecker", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": "JWT Authentication with Golang, Fiber, and GORM",
		})
	})

	// Authentication endpoints
	app.Route("/auth", func(router fiber.Router) {
		router.Post("/register", controller.User.SignUpUser)
		router.Post("/login", controller.User.SignInUser)
		router.Get("/logout", middleware.DeserializeUser, controller.User.LogoutUser)
		router.Get("/refresh", middleware.DeserializeUser, controller.User.RefreshToken)
	})
	logger_ := log.New(os.Stdout, "logger: ", log.Lshortfile)
	if config.EnableWebsocket {
		// log that we are starting the hub and pass the logger to it
		hub_ := hub.NewHub(brokers, logger_, nil)
		app.Route("/kafka", func(router fiber.Router) {
			router.Post("/send-message", middleware.DeserializeUser, controller.User.SendMessage)
			router.Get("/ws", middleware.DeserializeUser, websocket.New(func(c *websocket.Conn) {
				hub_.UpgradeWebSocket(c, logger_)
			}))
		})
	} else {
		app.Route("/kafka", func(router fiber.Router) {
			router.Post("/send-message", middleware.DeserializeUser, controller.User.SendMessage)
		})
	}

	// User details endpoint
	app.Get("/users/me", middleware.DeserializeUser, controller.User.GetMe)

	// Fallback route
	app.All("*", func(c *fiber.Ctx) error {
		path := c.Path()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "fail",
			"message": fmt.Sprintf("path: %v does not exist on this server", path),
		})
	})

	return app
}

// main is the entry point of the application
func main() {
	// Setup the fiber app
	var err error
	app, err = setupApp()
	if err != nil {
		log.Fatalf("failed to setup app: %v", err)
		return
	}

	// Setup the routes
	routes := setupRoutes(controller)

	// Mount the routes on /api
	app.Mount("/api", routes)

	// Get the port from the environment variable or use the default port
	port := os.Getenv("PORT")
	if port == "" {
		port = "8045"
	}

	// Start the server
	go func() {
		if err := app.Listen(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server listen: %s", err)
		}
	}()

	// Wait for a shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown the server gracefully
	log.Print("shutting down server...")
	if err := app.Shutdown(); err != nil {
		log.Fatal("server shutdown:", err)
	}

	log.Print("server exited")
}
