package middleware

import (
	"github.com/cploutarchou/go-kafka-rest/initializers"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"testing"
)

func TestNewMiddleware(t *testing.T) {
	config := &initializers.Config{} // Add real config if needed
	db := &gorm.DB{}                 // Add real db connection if needed

	middleware := NewMiddleware(config, db)

	assert.NotNil(t, middleware)
	assert.Equal(t, config, middleware.config)
	assert.Equal(t, db, middleware.db)
}
