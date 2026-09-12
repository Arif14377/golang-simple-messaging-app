package repository

import (
	"context"

	"github.com/Arif14377/golang-simple-messaging-app/app/models"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/database"
)

func InsertUserSession(ctx context.Context, session *models.UserSession) error {
	return database.DB.WithContext(ctx).Create(session).Error
}

func FindSessionByToken(ctx context.Context, tokenHash string) (models.UserSession, error) {
	var session models.UserSession
	err := database.DB.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&session).Error
	return session, err
}

func DeleteSessionByToken(ctx context.Context, tokenHash string) error {
	return database.DB.WithContext(ctx).Where("token_hash = ?", tokenHash).Delete(&models.UserSession{}).Error
}
