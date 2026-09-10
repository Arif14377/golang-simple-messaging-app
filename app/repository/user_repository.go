package repository

import (
	"context"

	"github.com/Arif14377/golang-simple-messaging-app/app/models"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/database"
)

func InsertNewUser(ctx context.Context, user *models.User) error {
	return database.DB.WithContext(ctx).Create(user).Error
}

func FindUserByUsername(ctx context.Context, username string) (models.User, error) {
	var user models.User
	err := database.DB.WithContext(ctx).First(&user, "username = ?", username).Error
	if err != nil {
		return user, err
	}

	return user, nil
}
