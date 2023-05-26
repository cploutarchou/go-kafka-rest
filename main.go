package main

import (
	"fmt"
	"log"
	"time"

	"github.com/cploutarchou/go-kafka-rest/controllers"
	"github.com/cploutarchou/go-kafka-rest/initializers"
	"github.com/cploutarchou/go-kafka-rest/middleware"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func init() {
	config, err := initializers.LoadConfig(".")
	if err != nil {
		log.Fatalf("Failed to load environment variables! \n%s", err.Error())
	}
	initializers.ConnectDB(config)
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
