package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainChatStorage "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/chatstorage"
)

var repo domainChatStorage.IChatStorageRepository

// SetRepository wires chat storage for per-device bucket resolution.
func SetRepository(r domainChatStorage.IChatStorageRepository) {
	repo = r
}

// ResolvedTarget describes where media for a device should be stored.
type ResolvedTarget struct {
	Driver       string
	Bucket       string
	DeviceFolder string
	Backend      Backend
}

// ResolveForDevice loads per-device config and picks local vs MiniCloud backend.
func ResolveForDevice(deviceID string) (*ResolvedTarget, error) {
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return nil, fmt.Errorf("device id is required for media storage")
	}
	folder := SanitizeDeviceFolder(deviceID, "")
	bucket := strings.TrimSpace(config.MiniCloudDefaultBucket)
	enabled := true
	if repo != nil {
		if cfg, err := repo.GetDeviceMediaStorage(deviceID); err != nil {
			return nil, err
		} else if cfg != nil {
			enabled = cfg.Enabled
			if strings.TrimSpace(cfg.Bucket) != "" {
				bucket = strings.TrimSpace(cfg.Bucket)
			}
			folder = SanitizeDeviceFolder(deviceID, cfg.FolderName)
		}
	}
	driver := strings.ToLower(strings.TrimSpace(config.MediaStorageDriver))
	if driver == "" {
		driver = "local"
	}
	useMiniCloud := driver == "minicloud" && enabled && strings.TrimSpace(config.MiniCloudBaseURL) != ""
	if useMiniCloud {
		if bucket == "" {
			return nil, fmt.Errorf("minicloud bucket is not configured for device %s", deviceID)
		}
		return &ResolvedTarget{
			Driver:       "minicloud",
			Bucket:       bucket,
			DeviceFolder: folder,
			Backend:      newMiniCloudBackend(),
		}, nil
	}
	return &ResolvedTarget{
		Driver:       "local",
		Bucket:       "",
		DeviceFolder: folder,
		Backend:      localBackend{},
	}, nil
}

// SaveMedia stores bytes for a device and returns a reference string (mc:... or local path).
func SaveMedia(ctx context.Context, deviceID string, data []byte, contentType string, ext string, kind MediaKind) (string, error) {
	target, err := ResolveForDevice(deviceID)
	if err != nil {
		return "", err
	}
	objectKey := BuildObjectKey(target.DeviceFolder, kind, ext)
	ref, err := target.Backend.Put(ctx, deviceID, target.Bucket, objectKey, data, contentType)
	if err != nil {
		return "", err
	}
	if target.Driver == "local" {
		return ref, nil
	}
	return ref, nil
}

// MaterializeToTempFile returns a local path suitable for os.Open (Chatwoot attachments).
func MaterializeToTempFile(ctx context.Context, mediaPath string) (path string, cleanup func(), err error) {
	mediaPath = strings.TrimSpace(mediaPath)
	if mediaPath == "" {
		return "", func() {}, fmt.Errorf("empty media path")
	}
	if !IsMediaRef(mediaPath) {
		if _, statErr := os.Stat(mediaPath); statErr != nil {
			return "", func() {}, statErr
		}
		return mediaPath, func() {}, nil
	}
	ref, err := ParseMC(mediaPath)
	if err != nil {
		return "", func() {}, err
	}
	data, _, err := newMiniCloudBackend().Get(ctx, ref)
	if err != nil {
		return "", func() {}, err
	}
	ext := filepath.Ext(ref.ObjectKey)
	f, err := os.CreateTemp("", "gowa-media-*"+ext)
	if err != nil {
		return "", func() {}, err
	}
	name := f.Name()
	if _, err = f.Write(data); err != nil {
		f.Close()
		os.Remove(name)
		return "", func() {}, err
	}
	if err = f.Close(); err != nil {
		os.Remove(name)
		return "", func() {}, err
	}
	return name, func() { os.Remove(name) }, nil
}

// ProxyObjectKey builds the wildcard path served under /devices/:id/media/*.
func ProxyObjectKey(deviceID, objectKeyWithinDevice string) string {
	target, err := ResolveForDevice(deviceID)
	if err != nil {
		return ""
	}
	key := strings.Trim(strings.ReplaceAll(objectKeyWithinDevice, "\\", "/"), "/")
	folder := target.DeviceFolder
	if folder == "" {
		return key
	}
	if key == "" || strings.HasPrefix(key, folder+"/") {
		return key
	}
	return folder + "/" + key
}

// PublicProxyURL builds the URL path segment for a stored mc: ref.
func PublicProxyURL(deviceID, mediaRef string) string {
	if !IsMediaRef(mediaRef) {
		return ""
	}
	ref, err := ParseMC(mediaRef)
	if err != nil {
		return ""
	}
	parts := strings.Split(ref.ObjectKey, "/")
	escaped := make([]string, len(parts))
	for i, p := range parts {
		escaped[i] = url.PathEscape(p)
	}
	return "/devices/" + url.PathEscape(deviceID) + "/media/" + strings.Join(escaped, "/")
}

// OpenForProxy streams media for GET /devices/:device_id/media/*.
func OpenForProxy(ctx context.Context, deviceID, wildcardPath string) (io.ReadCloser, string, error) {
	target, err := ResolveForDevice(deviceID)
	if err != nil {
		return nil, "", err
	}
	key := ObjectKeyFromProxyPath(target.DeviceFolder, wildcardPath)
	if key == "" {
		return nil, "", fmt.Errorf("invalid media path")
	}
	if target.Driver == "minicloud" {
		return target.Backend.OpenReader(ctx, &Ref{Bucket: target.Bucket, ObjectKey: key})
	}
	full := filepath.Join(config.PathMedia, filepath.FromSlash(key))
	f, err := os.Open(full)
	if err != nil {
		return nil, "", err
	}
	return f, "", nil
}
