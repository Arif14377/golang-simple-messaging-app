package repository

import (
	"context"

	"github.com/Arif14377/golang-simple-messaging-app/app/models"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/database"
)

func InsertNewUser(ctx context.Context, user *models.User) error {
	return database.DB.WithContext(ctx).Create(user).Error
}
