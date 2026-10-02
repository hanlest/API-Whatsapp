package validations

import (
	"context"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
)

type DeviceMediaStorageRequest struct {
	DeviceID   string
	Bucket     string
	FolderName string
	Enabled    *bool
}

func ValidateDeviceMediaStorageRequest(ctx context.Context, req *DeviceMediaStorageRequest) error {
	if req == nil || strings.TrimSpace(req.DeviceID) == "" {
		return pkgError.ValidationError("device_id is required")
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	driver := strings.ToLower(strings.TrimSpace(config.MediaStorageDriver))
	if enabled && driver == "minicloud" {
		bucket := strings.TrimSpace(req.Bucket)
		if bucket == "" {
			bucket = strings.TrimSpace(config.MiniCloudDefaultBucket)
		}
		if bucket == "" {
			return pkgError.ValidationError("bucket is required when media storage driver is minicloud (set per device or MINICLOUD_DEFAULT_BUCKET)")
		}
		if strings.TrimSpace(config.MiniCloudBaseURL) == "" {
			return pkgError.ValidationError("MINICLOUD_BASE_URL must be set when media storage driver is minicloud")
		}
	}
	return nil
}

func BuildDeviceMediaStorageFromRequest(req DeviceMediaStorageRequest, existing *domainChatStorage.DeviceMediaStorage) *domainChatStorage.DeviceMediaStorage {
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	} else if existing != nil {
		enabled = existing.Enabled
	}
	cfg := &domainChatStorage.DeviceMediaStorage{
		DeviceID:   req.DeviceID,
		Bucket:     strings.TrimSpace(req.Bucket),
		FolderName: strings.TrimSpace(req.FolderName),
		Enabled:    enabled,
	}
	if existing != nil {
		cfg.CreatedAt = existing.CreatedAt
	}
	return cfg
}
