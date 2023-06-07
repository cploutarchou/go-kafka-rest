package controllers

import (
	"github.com/cploutarchou/go-kafka-rest/initializers"
	"github.com/cploutarchou/go-kafka-rest/models"
	"github.com/golang-jwt/jwt"
	"time"
)

type AuthController struct{}

// CreateJWT creates a new JWT token for the given user.
func (a *AuthController) CreateJWT(user *models.User) (string, error) {
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
