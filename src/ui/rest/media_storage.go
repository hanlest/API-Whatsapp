package rest

import (
	"fmt"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/storage"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/validations"
	"github.com/gofiber/fiber/v3"
)

type MediaStorageHandler struct {
	ChatStorageRepo domainChatStorage.IChatStorageRepository
	DeviceManager   *whatsapp.DeviceManager
}

type mediaStorageRequest struct {
	Bucket     string `json:"bucket"`
	FolderName string `json:"folder_name"`
	Enabled    *bool  `json:"enabled"`
}

func mediaStorageView(cfg *domainChatStorage.DeviceMediaStorage) map[string]any {
	if cfg == nil {
		return nil
	}
	target, _ := storage.ResolveForDevice(cfg.DeviceID)
	view := map[string]any{
		"device_id":     cfg.DeviceID,
		"bucket":        cfg.Bucket,
		"folder_name":   cfg.FolderName,
		"enabled":       cfg.Enabled,
		"created_at":    cfg.CreatedAt,
		"updated_at":    cfg.UpdatedAt,
		"driver":        config.MediaStorageDriver,
		"device_folder": "",
		"effective_bucket": "",
	}
	if target != nil {
		view["device_folder"] = target.DeviceFolder
		view["effective_bucket"] = target.Bucket
	}
	return view
}

func (h *MediaStorageHandler) resolveDeviceID(c fiber.Ctx) (string, bool) {
	deviceID := pathDeviceID(c)
	if deviceID == "" || h.DeviceManager == nil {
		return "", false
	}
	_, resolvedID, err := h.DeviceManager.ResolveDevice(deviceID)
	if err != nil {
		return "", false
	}
	return strings.Clone(resolvedID), true
}

// ListMediaStorageConfigs GET /media-storage/configs
func (h *MediaStorageHandler) ListMediaStorageConfigs(c fiber.Ctx) error {
	if h.ChatStorageRepo == nil {
		return utils.ResponseError(c, "storage not available")
	}
	configs, err := h.ChatStorageRepo.ListDeviceMediaStorage()
	if err != nil {
		return utils.ResponseError(c, fmt.Sprintf("failed to list configs: %v", err))
	}
	views := make([]map[string]any, 0, len(configs))
	for _, cfg := range configs {
		views = append(views, mediaStorageView(cfg))
	}
	return c.JSON(utils.ResponseData{Status: 200, Code: "SUCCESS", Message: "Configuraciones de almacenamiento de medios", Results: views})
}

// GetMediaStorageConfig GET /devices/:device_id/media-storage
func (h *MediaStorageHandler) GetMediaStorageConfig(c fiber.Ctx) error {
	deviceID, ok := h.resolveDeviceID(c)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(utils.ResponseData{Status: fiber.StatusNotFound, Code: "DEVICE_NOT_FOUND", Message: "dispositivo no encontrado"})
	}
	cfg, err := h.ChatStorageRepo.GetDeviceMediaStorage(deviceID)
	if err != nil {
		return utils.ResponseError(c, fmt.Sprintf("failed to load config: %v", err))
	}
	if cfg == nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ResponseData{Status: fiber.StatusNotFound, Code: "CONFIG_NOT_FOUND", Message: "no hay configuración de medios para este dispositivo"})
	}
	return c.JSON(utils.ResponseData{Status: 200, Code: "SUCCESS", Message: "Configuración de almacenamiento de medios", Results: mediaStorageView(cfg)})
}

// UpsertMediaStorageConfig PUT /devices/:device_id/media-storage
func (h *MediaStorageHandler) UpsertMediaStorageConfig(c fiber.Ctx) error {
	deviceID, ok := h.resolveDeviceID(c)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(utils.ResponseData{Status: fiber.StatusNotFound, Code: "DEVICE_NOT_FOUND", Message: "dispositivo no encontrado"})
	}
	var req mediaStorageRequest
	if err := c.Bind().Body(&req); err != nil {
		return utils.ResponseError(c, "Invalid request body")
	}
	existing, err := h.ChatStorageRepo.GetDeviceMediaStorage(deviceID)
	if err != nil {
		return utils.ResponseError(c, fmt.Sprintf("failed to load existing config: %v", err))
	}
	validateReq := validations.DeviceMediaStorageRequest{
		DeviceID:   deviceID,
		Bucket:     req.Bucket,
		FolderName: req.FolderName,
		Enabled:    req.Enabled,
	}
	if err := validations.ValidateDeviceMediaStorageRequest(c.Context(), &validateReq); err != nil {
		return utils.ResponseError(c, err.Error())
	}
	cfg := validations.BuildDeviceMediaStorageFromRequest(validateReq, existing)
	if err := h.ChatStorageRepo.SaveDeviceMediaStorage(cfg); err != nil {
		return utils.ResponseError(c, fmt.Sprintf("failed to save config: %v", err))
	}
	return c.JSON(utils.ResponseData{Status: 200, Code: "SUCCESS", Message: "Configuración de medios guardada", Results: mediaStorageView(cfg)})
}

// DeleteMediaStorageConfig DELETE /devices/:device_id/media-storage
func (h *MediaStorageHandler) DeleteMediaStorageConfig(c fiber.Ctx) error {
	deviceID, ok := h.resolveDeviceID(c)
	if !ok {
		deviceID = pathDeviceID(c)
	}
	if deviceID == "" {
		return utils.ResponseError(c, "device_id is required")
	}
	if err := h.ChatStorageRepo.DeleteDeviceMediaStorage(deviceID); err != nil {
		return utils.ResponseError(c, fmt.Sprintf("failed to delete config: %v", err))
	}
	return c.JSON(utils.ResponseData{Status: 200, Code: "SUCCESS", Message: "Configuración de medios eliminada", Results: map[string]any{"device_id": deviceID}})
}

// ProxyDeviceMedia GET /devices/:device_id/media/*
func (h *MediaStorageHandler) ProxyDeviceMedia(c fiber.Ctx) error {
	deviceID, ok := h.resolveDeviceID(c)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(utils.ResponseData{Status: fiber.StatusNotFound, Code: "DEVICE_NOT_FOUND", Message: "dispositivo no encontrado"})
	}
	wildcard := c.Params("*")
	if wildcard == "" {
		wildcard = strings.TrimPrefix(c.Path(), "/devices/"+deviceID+"/media/")
	}
	rc, contentType, err := storage.OpenForProxy(c.Context(), deviceID, wildcard)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(utils.ResponseData{Status: fiber.StatusNotFound, Code: "MEDIA_NOT_FOUND", Message: "medio no encontrado: " + err.Error()})
	}
	defer rc.Close()
	if contentType != "" {
		c.Set("Content-Type", contentType)
	}
	c.Set("Cache-Control", "private, max-age=3600")
	return c.SendStream(rc)
}

// publicMediaFileURL returns a proxy URL for mc: refs or a static file URL for local paths.
func publicMediaFileURL(c fiber.Ctx, deviceID, filePath string) string {
	if storage.IsMediaRef(filePath) && config.MediaProxyEnabled {
		if proxyPath := storage.PublicProxyURL(deviceID, filePath); proxyPath != "" {
			return fmt.Sprintf("%s://%s%s%s", c.Scheme(), c.Host(), config.AppBasePath, proxyPath)
		}
	}
	return publicStaticFileURL(c, filePath)
}
