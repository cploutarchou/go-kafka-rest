package middleware

import (
	"github.com/cploutarchou/go-kafka-rest/initializers"
	"gorm.io/gorm"
)

type Middleware struct {
	config *initializers.Config
	db     *gorm.DB
}

func NewMiddleware(config *initializers.Config, db *gorm.DB) *Middleware {
	return &Middleware{
		config: config,
		db:     db,
	}
}
