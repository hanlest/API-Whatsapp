package middleware

import (
	"crypto/subtle"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest/helpers"
	"github.com/gofiber/fiber/v3"
)

// BaileysAPIKey enforces header x-api-key when config.AppAPIKey is set (Baileys API_KEY parity).
func BaileysAPIKey() fiber.Handler {
	return func(c fiber.Ctx) error {
		expected := strings.TrimSpace(config.AppAPIKey)
		if expected == "" {
			return c.Next()
		}

		provided := strings.TrimSpace(c.Get("x-api-key"))
		if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			return helpers.WriteBaileysError(c, fiber.StatusUnauthorized, "API key inválida o ausente")
		}
		return c.Next()
	}
}
