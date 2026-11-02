package models

import (
	"context"
	"time"

	"github.com/Arif14377/golang-simple-messaging-app/pkg/tracing"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type User struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	Username  string    `json:"username" gorm:"unique;type:varchar(20)" validate:"required,min=6,max=32"`
	FullName  string    `json:"fullName" gorm:"type:varchar(100);" validate:"required,min=6"`
	Password  string    `json:"password,omitempty" gorm:"type:varchar(255);" validate:"required,min=6"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}

func (l User) Validate(ctx context.Context) error {
	_, span := tracing.StartSpan(ctx, "User.Validate", "validate")
	defer span.End()

	return validate.Struct(l)
}

type UserSession struct {
	ID                  uint `gorm:"primarykey"`
	CreatedAt           time.Time
	UpdatedAt           time.Time
	UserID              uint      `json:"user_id" gorm:"type:int" validate:"required"`
	TokenHash           string    `json:"-" gorm:"size:64;index" validate:"required"`
	RefreshToken        string    `json:"refresh_token" gorm:"type:text" validate:"required"`
	RefreshTokenExpired time.Time `json:"-" validate:"required"`
}

func (l UserSession) Validate(ctx context.Context) error {
	_, span := tracing.StartSpan(ctx, "UserSession.Validate", "validate")
	defer span.End()

	return validate.Struct(l)
}
