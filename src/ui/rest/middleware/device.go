package middleware

import (
	"net/url"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

const DeviceIDHeader = "X-Device-Id"

// DeviceMiddleware fetches a device instance by header (preferred), path param, or query param
// and injects it into the context. It falls back to the default/only device for single-device mode.
func DeviceMiddleware(dm *whatsapp.DeviceManager) fiber.Handler {
	return func(c fiber.Ctx) error {
		// Allow non-device-scoped public endpoints (e.g., landing page) to pass through.
		path := strings.TrimSpace(c.Path())
		if path == "/" || path == "" || path == config.AppBasePath || path == config.AppBasePath+"/" {
			return c.Next()
		}

		if dm == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(utils.ResponseData{
				Status:  fiber.StatusServiceUnavailable,
				Code:    "DEVICE_MANAGER_UNAVAILABLE",
				Message: "El administrador de dispositivos no está inicializado",
				Results: nil,
			})
		}

		deviceID := strings.TrimSpace(c.Get(DeviceIDHeader))
		// URL-decode the header value to support non-ASCII characters
		if decoded, err := url.QueryUnescape(deviceID); err == nil {
			deviceID = decoded
		}
		if deviceID == "" {
			deviceID = strings.TrimSpace(c.Query("device_id"))
		}

		instance, resolvedID, err := dm.ResolveDevice(deviceID)
		if err != nil {
			// ResolveDevice returns an ID when provided but missing; use it for payload clarity.
			if resolvedID != "" || strings.TrimSpace(deviceID) != "" {
				return c.Status(fiber.StatusNotFound).JSON(utils.ResponseData{
					Status:  fiber.StatusNotFound,
					Code:    "DEVICE_NOT_FOUND",
					Message: "dispositivo no encontrado; creá uno en POST /devices o enviá un X-Device-Id válido",
					Results: map[string]string{"device_id": resolvedID},
				})
			}

			return c.Status(fiber.StatusBadRequest).JSON(utils.ResponseData{
				Status:  fiber.StatusBadRequest,
				Code:    "DEVICE_ID_REQUIRED",
				Message: "device_id es obligatorio (header X-Device-Id o query device_id)",
				Results: nil,
			})
		}

		c.Locals("device_id", resolvedID)
		c.Locals("device", instance)
		c.SetContext(whatsapp.ContextWithDevice(c.Context(), instance))
		return c.Next()
	}
}
