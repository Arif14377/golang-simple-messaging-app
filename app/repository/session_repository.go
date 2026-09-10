package repository

import (
	"context"

	"github.com/Arif14377/golang-simple-messaging-app/app/models"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/database"
)

func InsertUserSession(ctx context.Context, session *models.UserSession) error {
	return database.DB.WithContext(ctx).Create(session).Error
}
