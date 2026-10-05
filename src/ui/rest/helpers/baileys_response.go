package helpers

import (
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/gofiber/fiber/v3"
)

type BaileysEnvelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    *BaileysData `json:"data,omitempty"`
}

type BaileysData struct {
	ID *string `json:"id"`
}

func WriteBaileysSuccess(c fiber.Ctx, message string, messageID string) error {
	var idPtr *string
	if messageID != "" {
		idPtr = &messageID
	}
	return c.Status(fiber.StatusOK).JSON(BaileysEnvelope{
		Success: true,
		Message: message,
		Data:    &BaileysData{ID: idPtr},
	})
}

func WriteBaileysError(c fiber.Ctx, status int, message string) error {
	return c.Status(status).JSON(BaileysEnvelope{
		Success: false,
		Message: message,
	})
}

func WriteBaileysUsecaseError(c fiber.Ctx, err error) error {
	if err == nil {
		return WriteBaileysError(c, fiber.StatusInternalServerError, "Error interno")
	}
	if ge, ok := err.(pkgError.GenericError); ok {
		return WriteBaileysError(c, ge.StatusCode(), ge.Error())
	}
	return WriteBaileysError(c, fiber.StatusInternalServerError, err.Error())
}
