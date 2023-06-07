package utils

import (
	"errors"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
)

// HTTPError is a struct that represents an error response.
type HTTPError struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// RespondError creates an HTTPError and responds with the given error status and message.
func RespondError(c *fiber.Ctx, status int, message string) error {
	err := HTTPError{
		Status:  "error",
		Message: message,
	}
	return c.Status(status).JSON(err)
}

// GenerateHashedPassword generates a hashed password from the given password.
// It returns an error if there's a failure generating the password.
func GenerateHashedPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", errors.New("password must be at least 8 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}
