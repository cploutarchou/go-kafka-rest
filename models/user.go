package models

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

// User holds user related properties
type User struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primary_key"`
	Name      string    `gorm:"type:varchar(100);not null"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Password  string    `gorm:"type:varchar(100);not null"`
	Role      string    `gorm:"type:varchar(50);default:'user';not null"`
	Provider  string    `gorm:"type:varchar(50);default:'local';not null"`
	Verified  bool      `gorm:"not null;default:false"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	db        *gorm.DB
}

// SignUpInput holds user signup properties
type SignUpInput struct {
	Name            string `json:"name" validate:"required"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=8"`
	PasswordConfirm string `json:"passwordConfirm" validate:"required,min=8"`
}

// SignInInput holds user signin properties
type SignInInput struct {
	Email    string `json:"email"  validate:"required,email"`
	Password string `json:"password"  validate:"required"`
}

// UserResponse holds user response properties
type UserResponse struct {
	ID        uuid.UUID `json:"id,omitempty"`
	Name      string    `json:"name,omitempty"`
	Email     string    `json:"email,omitempty"`
	Role      string    `json:"role,omitempty"`
	Provider  string    `json:"provider"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// FilterUserRecord returns a filtered User response
func FilterUserRecord(user *User) UserResponse {
	return UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		Provider:  user.Provider,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

var validate = validator.New()

type ErrorResponse struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value,omitempty"`
}

// ValidateStruct validates a struct and returns an ErrorResponse array
func ValidateStruct[T any](payload T) []*ErrorResponse {
	var errors []*ErrorResponse
	err := validate.Struct(payload)
	if err != nil {
		for _, err := range err.(validator.ValidationErrors) {
			var element ErrorResponse
			element.Field = err.StructNamespace()
			element.Tag = err.Tag()
			element.Value = err.Param()
			errors = append(errors, &element)
		}
	}
	return errors
}

// GetByEmail returns a user from database by email
func (u *User) GetByEmail(email string) (*User, error) {
	var user User
	if result := u.db.Where("email = ?", email).First(&user); result.Error != nil {
		return nil, fmt.Errorf("user with email %s not found", email)
	}
	return &user, nil
}

// GetByID returns a user from database by ID
func (u *User) GetByID(id uuid.UUID) (*User, error) {
	var user User
	if result := u.db.Where("id = ?", id).First(&user); result.Error != nil {
		return nil, fmt.Errorf("user with id %s not found", id)
	}
	return &user, nil
}

// SetDB sets the DB connection for a user instance
func (u *User) SetDB(db *gorm.DB) {
	u.db = db
}
