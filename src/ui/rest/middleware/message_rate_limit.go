package middleware

import (
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest/helpers"
	"github.com/gofiber/fiber/v3"
)

type rateBucket struct {
	count   int
	windowStart time.Time
}

var (
	messageRateMu sync.Mutex
	messageRates  = map[string]*rateBucket{}
)

// MessageRateLimiter limits outbound message-style API calls per client IP (Baileys messageRateLimiter parity).
func MessageRateLimiter() fiber.Handler {
	return func(c fiber.Ctx) error {
		limit := config.AppMessageRateLimit
		if limit <= 0 {
			return c.Next()
		}

		key := c.IP()
		now := time.Now()

		messageRateMu.Lock()
		b, ok := messageRates[key]
		if !ok || now.Sub(b.windowStart) >= time.Minute {
			messageRates[key] = &rateBucket{count: 1, windowStart: now}
			messageRateMu.Unlock()
			return c.Next()
		}
		if b.count >= limit {
			messageRateMu.Unlock()
			return helpers.WriteBaileysError(c, fiber.StatusTooManyRequests, "Demasiadas solicitudes, intente más tarde")
		}
		b.count++
		messageRateMu.Unlock()
		return c.Next()
	}
}
