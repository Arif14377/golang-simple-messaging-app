package dto

import (
	"time"

	"github.com/Arif14377/golang-simple-messaging-app/app/models"
)

// MessageResponse adalah bentuk pesan yang dikirim ke client, baik lewat
// WebSocket maupun endpoint riwayat. Field `from` dan `message` dipertahankan
// agar kontrak lama dengan frontend tetap kompatibel.
type MessageResponse struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"roomId"`
	From      string    `json:"from"`
	UserID    uint      `json:"userId"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

func NewMessageResponse(message models.Message) MessageResponse {
	return MessageResponse{
		ID:        message.ID.Hex(),
		RoomID:    message.RoomID,
		From:      message.Username,
		UserID:    message.UserID,
		Message:   message.Content,
		CreatedAt: message.CreatedAt,
	}
}

func NewMessageResponses(messages []models.Message) []MessageResponse {
	result := make([]MessageResponse, 0, len(messages))
	for _, message := range messages {
		result = append(result, NewMessageResponse(message))
	}

	return result
}
