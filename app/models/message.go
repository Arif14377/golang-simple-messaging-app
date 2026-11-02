package models

import (
	"context"
	"time"

	"github.com/Arif14377/golang-simple-messaging-app/pkg/tracing"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// Message adalah dokumen chat yang disimpan di MongoDB.
// Semua field identitas (user id, username, waktu) diisi oleh server agar
// tidak bisa dipalsukan client.
type Message struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	RoomID    string        `bson:"room_id" json:"roomId" validate:"required,max=64"`
	UserID    uint          `bson:"user_id" json:"userId" validate:"required"`
	Username  string        `bson:"username" json:"username" validate:"required,max=32"`
	Content   string        `bson:"content" json:"content" validate:"required,max=2000"`
	CreatedAt time.Time     `bson:"created_at" json:"createdAt" validate:"required"`
}

// Validate memastikan dokumen pesan memiliki data wajib yang valid.
func (m Message) Validate(ctx context.Context) error {
	_, span := tracing.StartSpan(ctx, "Message.Validate", "validate")
	defer span.End()

	return validate.Struct(m)
}
