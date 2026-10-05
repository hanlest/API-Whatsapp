package middleware

import (
	"context"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
)

// DefaultRequestTimeout is the default timeout for API requests.
// This prevents indefinite blocking when WhatsApp servers are slow or unresponsive.
const DefaultRequestTimeout = 45 * time.Second

// RequestTimeout adds a deadline to all incoming HTTP requests.
// If the handler doesn't complete within the timeout, the context is cancelled
// and whatsmeow SDK calls will return context.DeadlineExceeded.
func RequestTimeout(timeout time.Duration) fiber.Handler {
	const statusPublishTimeout = 15 * time.Minute
	return func(c fiber.Ctx) error {
		effective := timeout
		path := c.Path()
		if strings.HasSuffix(path, "/send/status") || strings.HasSuffix(path, "/whatsapp/send-status") {
			effective = statusPublishTimeout
		}
		ctx, cancel := context.WithTimeout(c.Context(), effective)
		defer cancel()
		c.SetContext(ctx)
		return c.Next()
	}
}
