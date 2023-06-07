package utils

import (
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestRespondError(t *testing.T) {
	// Initialize a new Fiber app
	app := fiber.New()

	// Setup a test route
	app.Get("/test", func(c *fiber.Ctx) error {
		// Define error status and message
		status := 400
		message := "Test error message"

		// Call RespondError
		return RespondError(c, status, message)
	})

	// Perform a request to the test route
	resp, err := app.Test(httptest.NewRequest("GET", "/test", nil))
	if err != nil {
		t.Fatal(err)
	}

	// Assert status code
	assert.Equal(t, 400, resp.StatusCode)

	// Read the response body
	body, _ := io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Assert response body
	expectedBody := `{"status":"error","message":"Test error message"}`
	assert.Equal(t, expectedBody, string(body))
}

func TestGenerateHashedPassword(t *testing.T) {
	// Test valid password
	password := "testpassword123"
	hashed, err := GenerateHashedPassword(password)
	assert.NoError(t, err)

	// The hashed password should not be the same as the original password
	assert.NotEqual(t, password, hashed)

	// The hashed password should be correct when compared with the original password
	err = bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password))
	assert.NoError(t, err)

	// Test password that's too short
	shortPassword := "abc"
	_, err = GenerateHashedPassword(shortPassword)
	assert.Error(t, err)
	assert.Equal(t, "password must be at least 8 characters", err.Error())
}
