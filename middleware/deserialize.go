package middleware

import (
	"fmt"
	"strings"

	"github.com/cploutarchou/go-kafka-rest/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

func (m *Middleware) DeserializeUser(c *fiber.Ctx) error {
	tokenString := getTokenString(c)
	if tokenString == "" {
		return sendErrorResponse(c, fiber.StatusUnauthorized, "You are not logged in")
	}

	token, err := m.parseToken(tokenString, m.config.JwtSecret)
	if err != nil {
		return sendErrorResponse(c, fiber.StatusUnauthorized, fmt.Sprintf("Invalid token: %v", err))
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return sendErrorResponse(c, fiber.StatusUnauthorized, "Invalid token claim")
	}

	userIDClaim, ok := claims["sub"].(string)
	if !ok {
		return sendErrorResponse(c, fiber.StatusUnauthorized, "Invalid token claim")
	}

	user, err := m.getUserByID(userIDClaim)
	if err != nil {
		return sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to retrieve user")
	}

	if user == nil {
		return sendErrorResponse(c, fiber.StatusForbidden, "The user belonging to this token no longer exists")
	}

	c.Locals("user", models.FilterUserRecord(user))
	return c.Next()
}

func (m *Middleware) parseToken(tokenString string, jwtSecret string) (*jwt.Token, error) {
	if strings.HasPrefix(tokenString, "Bearer ") {
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	}

	token, err := jwt.Parse(tokenString, func(jwtToken *jwt.Token) (interface{}, error) {
		if _, ok := jwtToken.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %s", jwtToken.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	return token, err
}

func (m *Middleware) getUserByID(userID string) (*models.User, error) {
	var user models.User
	err := m.db.First(&user, "id = ?", userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func getTokenString(c *fiber.Ctx) string {
	if token := c.Cookies("token"); token != "" {
		return token
	}
	return c.Get("Authorization")
}

func sendErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{"status": "error", "message": message})
}
