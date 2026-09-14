package dto

import (
	"github.com/Arif14377/golang-simple-messaging-app/app/models"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type RegisterResponse struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
}

func NewRegisterResponse(data models.User) *RegisterResponse {
	return &RegisterResponse{
		UserID:   data.ID,
		Username: data.Username,
		FullName: data.FullName,
	}
}

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

func (l LoginRequest) Validate() error {
	return validate.Struct(l)
}

type LoginResponse struct {
	Username     string `json:"username" `
	FullName     string `json:"fullName" `
	Token        string `json:"token" `
	RefreshToken string `json:"refreshToken" `
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

func (l RefreshTokenRequest) Validate() error {
	return validate.Struct(l)
}

type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
}
