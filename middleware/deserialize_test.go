package middleware

import (
	"github.com/cploutarchou/go-kafka-rest/initializers"
	"testing"

	"github.com/golang-jwt/jwt"
	"github.com/stretchr/testify/assert"
)

func TestParseToken(t *testing.T) {
	secret := "test_secret"

	m := &Middleware{
		config: &initializers.Config{
			JwtSecret: secret,
		},
	}

	// Creating a valid JWT token
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)
	claims["sub"] = "1234" // user id
	claims["name"] = "Test User"

	tokenString, _ := token.SignedString([]byte(secret))

	_, err := m.parseToken(tokenString, m.config.JwtSecret)

	assert.NoError(t, err)
}
