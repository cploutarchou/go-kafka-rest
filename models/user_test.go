package models

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestValidateStruct(t *testing.T) {
	// Create a valid SignUpInput instance
	signUpInput := &SignUpInput{
		Name:            "Test User",
		Email:           "testuser@test.com",
		Password:        "password123",
		PasswordConfirm: "password123",
	}

	// Validate the struct
	errors := ValidateStruct(signUpInput)

	// Check that no validation errors are returned
	assert.Equal(t, 0, len(errors))

	// Create an invalid SignUpInput instance
	invalidSignUpInput := &SignUpInput{
		Name:            "",
		Email:           "invalid email",
		Password:        "short",
		PasswordConfirm: "short",
	}

	// Validate the struct
	errors = ValidateStruct(invalidSignUpInput)

	// Check that validation errors are returned
	assert.Greater(t, len(errors), 0)
}
func TestGetByEmail(t *testing.T) {
	// Create a new sqlmock instance
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	// Use gorm with the sqlmock database
	gormDB, _ := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})

	// Create a user instance
	user := &User{
		ID:        uuid.New(),
		Name:      "Test User",
		Email:     "testuser@test.com",
		Password:  "password123",
		Role:      "user",
		Provider:  "local",
		Verified:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Set the DB connection for the user instance
	user.SetDB(gormDB)

	// Expect a query to select from the users table
	mock.ExpectQuery("^SELECT (.+) FROM \"users\" WHERE (.+) LIMIT 1").
		WithArgs(user.Email).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "email", "password", "role", "provider", "verified", "created_at", "updated_at"}).
			AddRow(user.ID, user.Name, user.Email, user.Password, user.Role, user.Provider, user.Verified, user.CreatedAt, user.UpdatedAt))

	// Call the GetByEmail method
	result, err := user.GetByEmail(user.Email)

	// Check that no error was returned
	assert.NoError(t, err)

	// Check that the result matches the expected user
	assert.Equal(t, user.ID, result.ID)
	assert.Equal(t, user.Name, result.Name)
	assert.Equal(t, user.Email, result.Email)
	assert.Equal(t, user.Password, result.Password)
	assert.Equal(t, user.Role, result.Role)
	assert.Equal(t, user.Provider, result.Provider)
	assert.Equal(t, user.Verified, result.Verified)

}
