package domain

import (
	"time"

	"github.com/google/uuid"
	apperrors "github.com/microservices/shared/errors"
	"golang.org/x/crypto/bcrypt"
)

type Role string

const (
	RoleAdmin    Role = "admin"
	RoleCustomer Role = "customer"
)

type User struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         Role      `json:"role"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUserInput struct {
	Name     string `json:"name"     binding:"required,min=2,max=100"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role     Role   `json:"role"`
}

type UpdateUserInput struct {
	Name  string `json:"name"  binding:"omitempty,min=2,max=100"`
	Email string `json:"email" binding:"omitempty,email"`
}

type LoginInput struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func NewUser(input CreateUserInput) (*User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperrors.Wrap(err, "hash password")
	}

	role := input.Role
	if role == "" {
		role = RoleCustomer
	}

	now := time.Now()
	return &User{
		ID:           uuid.NewString(),
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
		Role:         role,
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

func (u *User) Update(input UpdateUserInput) {
	if input.Name != "" {
		u.Name = input.Name
	}
	if input.Email != "" {
		u.Email = input.Email
	}
	u.UpdatedAt = time.Now()
}
