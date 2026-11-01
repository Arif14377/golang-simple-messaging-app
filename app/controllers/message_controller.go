package controllers

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/Arif14377/golang-simple-messaging-app/app/repository"
	"github.com/Arif14377/golang-simple-messaging-app/app/ws"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/dto"
	"github.com/Arif14377/golang-simple-messaging-app/pkg/response"
	"github.com/gofiber/fiber/v3"
)

// MessageHistory mengembalikan riwayat pesan sebuah room, terurut dari yang
// paling lama ke paling baru. Query opsional:
//   - room   (default "lobby")
//   - limit  (default 50, maksimal 100)
//   - before (RFC3339, untuk pagination ke pesan yang lebih lama)
func MessageHistory(ctx fiber.Ctx) error {
	roomID := strings.TrimSpace(ctx.Query("room", ws.DefaultRoomID))
	if roomID == "" {
		roomID = ws.DefaultRoomID
	}

	limit := repository.DefaultHistoryLimit
	if rawLimit := strings.TrimSpace(ctx.Query("limit")); rawLimit != "" {
		parsed, err := strconv.ParseInt(rawLimit, 10, 64)
		if err != nil || parsed <= 0 {
			return response.SendFailureResponse(ctx, fiber.StatusBadRequest, "Invalid limit.")
		}
		limit = parsed
	}

	var before time.Time
	if rawBefore := strings.TrimSpace(ctx.Query("before")); rawBefore != "" {
		parsed, err := time.Parse(time.RFC3339, rawBefore)
		if err != nil {
			return response.SendFailureResponse(ctx, fiber.StatusBadRequest, "Invalid before timestamp.")
		}
		before = parsed
	}

	cctx, cancel := context.WithTimeout(ctx.Context(), 5*time.Second)
	defer cancel()

	messages, err := repository.FindMessagesByRoom(cctx, roomID, limit, before)
	if err != nil {
		log.Printf("Failed to find messages: %v\n", err)
		return response.SendFailureResponse(ctx, fiber.StatusInternalServerError, "Failed to get message history")
	}

	return response.SendSuccessResponse(ctx, fiber.StatusOK, "Message history", dto.NewMessageResponses(messages))
}
