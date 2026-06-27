package auth

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	Email    string
	Password string

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserResponse struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
}

type RegisterRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func FromRegisterRequest(req *RegisterRequest, hashedPassword string) User {
	return User{
		ID:       uuid.New(),
		Email:    req.Email,
		Password: hashedPassword,
	}
}

func ToUserResponse(u *User) *UserResponse {
	return &UserResponse{
		ID:    u.ID,
		Email: u.Email,
	}
}
