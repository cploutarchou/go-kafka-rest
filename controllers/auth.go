package controllers

import (
	"fmt"
	"strings"
	"time"

	"github.com/cploutarchou/go-kafka-rest/initializers"
	"github.com/cploutarchou/go-kafka-rest/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

// RespondError responds with the given error status and message.
func RespondError(c *fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(fiber.Map{
		"status":  "error",
		"message": message,
	})
}

// GenerateHashedPassword generates a hashed password from the given password.
func GenerateHashedPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// CreateJWT creates a new JWT token for the given user.
func CreateJWT(user *models.User) (string, error) {
	config, _ := initializers.LoadConfig(".")
	tokenByte := jwt.New(jwt.SigningMethodHS256)
	now := time.Now().UTC()
	claims := tokenByte.Claims.(jwt.MapClaims)
	claims["sub"] = user.ID
	claims["exp"] = now.Add(config.JwtExpiresIn).Unix()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()
	return tokenByte.SignedString([]byte(config.JwtSecret))
}

// SignUpUser creates a new user with the given name, email and password.
func SignUpUser(c *fiber.Ctx) error {
	var payload *models.SignUpInput

	if err := c.BodyParser(&payload); err != nil {
		return RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	errors := models.ValidateStruct(payload)
	if errors != nil {
		return RespondError(c, fiber.StatusBadRequest, fmt.Sprint(errors))
	}

	if payload.Password != payload.PasswordConfirm {
		return RespondError(c, fiber.StatusBadRequest, "Passwords do not match")
	}

	hashedPassword, err := GenerateHashedPassword(payload.Password)
	if err != nil {
		return RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	newUser := models.User{
		Name:     payload.Name,
		Email:    strings.ToLower(payload.Email),
		Password: hashedPassword,
		Photo:    &payload.Photo,
	}

	result := initializers.DB.Create(&newUser)
	if result.Error != nil {
		if strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
			return RespondError(c, fiber.StatusConflict, "User with that email already exists")
		}
		return RespondError(c, fiber.StatusBadGateway, "Something bad happened")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": "success", "data": fiber.Map{"user": models.FilterUserRecord(&newUser)}})
}

// SignUpUser creates a new user with the given name, email and password.
func SignInUser(c *fiber.Ctx) error {
	var payload *models.SignInInput

	if err := c.BodyParser(&payload); err != nil {
		return RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	errors := models.ValidateStruct(payload)
	if errors != nil {
		return RespondError(c, fiber.StatusBadRequest, fmt.Sprint(errors))
	}

	var user models.User
	result := initializers.DB.First(&user, "email = ?", strings.ToLower(payload.Email))
	if result.Error != nil {
		return RespondError(c, fiber.StatusNotFound, "User with that email does not exist")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password)); err != nil {
		return RespondError(c, fiber.StatusUnauthorized, "Invalid password")
	}

	token, err := CreateJWT(&user)
	if err != nil {
		return RespondError(c, fiber.StatusBadGateway, "Something bad happened")
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "data": fiber.Map{"token": token}})
}

// LogoutUser logs out the currently logged in user.
func LogoutUser(c *fiber.Ctx) error {
	user := c.Locals("user").(*jwt.Token)
	claims := user.Claims.(jwt.MapClaims)
	claims["exp"] = time.Now().Unix()
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Logged out successfully"})
}