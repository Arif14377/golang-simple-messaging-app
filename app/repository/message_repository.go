package repository

import (
	"context"
	"slices"
	"time"

	"github.com/Arif14377/golang-simple-messaging-app/app/models"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/database"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

const (
	// DefaultHistoryLimit adalah jumlah pesan default saat mengambil riwayat.
	DefaultHistoryLimit int64 = 50
	// MaxHistoryLimit membatasi jumlah pesan yang boleh diminta client.
	MaxHistoryLimit int64 = 100
)

// InsertMessage menyimpan satu pesan ke MongoDB dan mengisi ID hasil insert.
func InsertMessage(ctx context.Context, message *models.Message) error {
	ctx, span := startMongoSpan(ctx, "InsertMessage", "insert")
	defer span.End()

	result, err := database.MessageCollection().InsertOne(ctx, message)
	if err != nil {
		return err
	}

	if id, ok := result.InsertedID.(bson.ObjectID); ok {
		message.ID = id
	}

	return nil
}

// FindMessagesByRoom mengambil riwayat pesan sebuah room, terbaru lebih dulu
// di query lalu dikembalikan dalam urutan lama → baru agar siap dirender UI.
// before dipakai sebagai cursor: hanya pesan yang lebih lama dari waktu itu.
func FindMessagesByRoom(ctx context.Context, roomID string, limit int64, before time.Time) ([]models.Message, error) {
	ctx, span := startMongoSpan(ctx, "FindMessagesByRoom", "find")
	defer span.End()

	if limit <= 0 {
		limit = DefaultHistoryLimit
	}
	if limit > MaxHistoryLimit {
		limit = MaxHistoryLimit
	}

	filter := bson.M{"room_id": roomID}
	if !before.IsZero() {
		filter["created_at"] = bson.M{"$lt": before}
	}

	findOptions := options.Find().
		SetSort(bson.D{
			{Key: "created_at", Value: -1},
			{Key: "_id", Value: -1},
		}).
		SetLimit(limit)

	cursor, err := database.MessageCollection().Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	messages := make([]models.Message, 0, limit)
	if err := cursor.All(ctx, &messages); err != nil {
		return nil, err
	}

	// query mengambil dari yang terbaru, UI butuh urutan kronologis.
	slices.Reverse(messages)

	return messages, nil
}
