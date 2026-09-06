package dto

import "github.com/Arif14377/golang-simple-messaging-app/app/models"

type UserRegister struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
}

func NewUserRegister(data models.User) *UserRegister {
	return &UserRegister{
		UserID:   data.ID,
		Username: data.Username,
		FullName: data.FullName,
	}
}
