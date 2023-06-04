package controllers

import (
	"fmt"
	"github.com/cploutarchou/go-kafka-rest/initializers"
	"github.com/cploutarchou/go-kafka-rest/models"
	"github.com/cploutarchou/go-kafka-rest/types"
	"github.com/cploutarchou/go-kafka-rest/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"net/http"
	"strings"
	"time"
)

type UserController struct {
	Model          models.User
	DB             *gorm.DB
	AuthController AuthController
}

func (u *UserController) GetMe(c *fiber.Ctx) error {
	user := c.Locals("user").(models.UserResponse)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"user": user}})
}

func (u *UserController) RefreshToken(ctx *fiber.Ctx) error {
	user := ctx.Locals("user").(models.UserResponse)

	fmt.Println(user.Email)

	user_, err := u.Model.GetByEmail(user.Email)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Failed to retrieve user"})
	}
	token, err := u.AuthController.CreateJWT(user_)
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "error", "message": "Something went wrong"})
	}

	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"token": token}})
}

func (u *UserController) SignUpUser(c *fiber.Ctx) error {
	var payload *models.SignUpInput

	if err := c.BodyParser(&payload); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	errors := models.ValidateStruct(payload)
	if errors != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, fmt.Sprint(errors))
	}

	if payload.Password != payload.PasswordConfirm {
		return utils.RespondError(c, fiber.StatusBadRequest, "Passwords do not match")
	}

	hashedPassword, err := utils.GenerateHashedPassword(payload.Password)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	newUser := models.User{
		Name:     payload.Name,
		Email:    strings.ToLower(payload.Email),
		Password: hashedPassword,
	}

	result := initializers.DB.Create(&newUser)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
			return utils.RespondError(c, fiber.StatusConflict, "User with that email already exists")
		}
		return utils.RespondError(c, fiber.StatusBadGateway, "Something bad happened")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": fiber.Map{"user": models.FilterUserRecord(&newUser)}})
}

func (u *UserController) SignInUser(c *fiber.Ctx) error {
	var payload *models.SignInInput

	if err := c.BodyParser(&payload); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	errors := models.ValidateStruct(payload)
	if errors != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, fmt.Sprint(errors))
	}

	var user models.User
	result := initializers.DB.First(&user, "email = ?", strings.ToLower(payload.Email))
	if result.Error != nil {
		return utils.RespondError(c, fiber.StatusNotFound, "User with that email does not exist")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		return utils.RespondError(c, fiber.StatusUnauthorized, "Invalid password")
	}

	token, err := u.AuthController.CreateJWT(&user)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadGateway, "Something bad happened")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"token": token}})
}

func (u *UserController) LogoutUser(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Unix()
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Logged out successfully"})
}

func (u *UserController) SendMessage(c *fiber.Ctx) error {
	var messagePayload types.MessagePayload
	err := c.BodyParser(&messagePayload)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"status":  "fail",
			"message": "Invalid request payload",
		})
	}

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

	mutex.Lock()
	messageQueue = append(messageQueue, messagePayload)
	mutex.Unlock()

	workerPool <- struct{}{}
	wg.Add(1)

	go func() {
		defer func() {
			<-workerPool
			wg.Done()
		}()

		for {
			mutex.Lock()
			if len(messageQueue) == 0 {
				mutex.Unlock()
				break
			}
			message := messageQueue[0]
			messageQueue = messageQueue[1:]
			mutex.Unlock()

			producer.SendMessageAsync(message.Topic, message.Key, message.Data)
			log.Printf("Kafka message produced! Topic: %s\n", message.Topic)
		}
	}()

	return c.Status(http.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Message received and added to the processing queue",
	})
}
