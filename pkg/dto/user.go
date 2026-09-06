package dto

import "github.com/Arif14377/golang-simple-messaging-app/app/models"

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
