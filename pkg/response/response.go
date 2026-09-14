package response

import "github.com/gofiber/fiber/v3"

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type FailureResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func SendSuccessResponse(ctx fiber.Ctx, statusCode int, message string, data any) error {
	return ctx.Status(statusCode).JSON(SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SendFailureResponse(ctx fiber.Ctx, statusCode int, message string) error {
	return ctx.Status(statusCode).JSON(FailureResponse{
		Success: false,
		Message: message,
	})
}
