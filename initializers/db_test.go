package initializers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConnectDB(t *testing.T) {

	ConnectDB(GetConfig())

	// Check that DB is not nil after connection
	assert.NotNil(t, DB)

	// Perform any other checks to validate successful connection
	// You may want to execute some queries and check their result
}

func TestGetDB(t *testing.T) {
	// Assuming ConnectDB has been successfully called before
	db := GetDB()

	// Check that GetDB returns the DB instance
	assert.Equal(t, DB, db)

	// Perform any other checks to validate that the DB instance works
	// You may want to execute some queries and check their result
}
