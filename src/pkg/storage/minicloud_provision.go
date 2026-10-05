package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
)

func isBucketNotConfiguredError(uploadErr error) bool {
	if uploadErr == nil {
		return false
	}
	msg := strings.ToLower(uploadErr.Error())
	return strings.Contains(msg, "400") &&
		(strings.Contains(msg, "bucket_not_configured") ||
			strings.Contains(msg, "sin almacenamientos") ||
			strings.Contains(msg, "bucket sin"))
}

// resolveMiniCloudStorageID picks the MiniCloud storage backend id (storages.json entry).
// Default: sanitized device slot name; optional global override via MINICLOUD_DEFAULT_STORAGE_ID.
func resolveMiniCloudStorageID(deviceID string) string {
	if id := strings.TrimSpace(config.MiniCloudDefaultStorageID); id != "" {
		return SanitizeDeviceFolder(id, "")
	}
	folderOverride := ""
	if repo != nil {
		if cfg, err := repo.GetDeviceMediaStorage(deviceID); err == nil && cfg != nil {
			folderOverride = cfg.FolderName
		}
	}
	return SanitizeDeviceFolder(deviceID, folderOverride)
}

// ensureMiniCloudBucket registers the bucket and binds the device-named storage as primary.
func ensureMiniCloudBucket(ctx context.Context, m *miniCloudBackend, deviceID, bucket string) error {
	if !config.MiniCloudAutoCreateBucket {
		return fmt.Errorf("minicloud auto-create bucket is disabled")
	}
	storageID := resolveMiniCloudStorageID(deviceID)
	if storageID == "" {
		return fmt.Errorf("could not resolve storage id for device %q", deviceID)
	}

	if err := m.adminJSON(ctx, http.MethodPost, "/admin/api/storage/buckets", map[string]string{
		"name": bucket,
	}, []int{http.StatusCreated, http.StatusConflict}); err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}

	existing, err := m.listBucketPrimaryStorageIDs(ctx, bucket)
	if err != nil {
		return err
	}
	if len(existing) > 0 {
		for _, id := range existing {
			if id == storageID {
				return m.ensureStoragePrimary(ctx, storageID, bucket)
			}
		}
		// Bucket ya tiene principal(es); no crear otro por cada dispositivo registrado.
		return nil
	}

	if err := m.ensureStoragePrimary(ctx, storageID, bucket); err != nil {
		return err
	}
	return nil
}

func (m *miniCloudBackend) listBucketPrimaryStorageIDs(ctx context.Context, bucket string) ([]string, error) {
	q := url.QueryEscape(bucket)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.baseURL+"/admin/api/storage/storages?bucket="+q, nil)
	if err != nil {
		return nil, err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return nil, fmt.Errorf("list storages for bucket: %s", strings.TrimSpace(string(b)))
	}
	var payload struct {
		Storages []struct {
			ID          string `json:"id"`
			StorageType string `json:"storageType"`
			Enabled     bool   `json:"enabled"`
		} `json:"storages"`
		Backends []struct {
			ID          string `json:"id"`
			StorageType string `json:"storageType"`
			Enabled     bool   `json:"enabled"`
		} `json:"backends"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	list := payload.Storages
	if len(list) == 0 {
		list = payload.Backends
	}
	var ids []string
	for _, s := range list {
		if !s.Enabled {
			continue
		}
		st := strings.ToLower(strings.TrimSpace(s.StorageType))
		if st != "" && st != "primary" {
			continue
		}
		if id := strings.TrimSpace(s.ID); id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

func (m *miniCloudBackend) ensureStoragePrimary(ctx context.Context, storageID, bucket string) error {
	patchPath := "/admin/api/storage/storages/" + url.PathEscape(storageID)
	err := m.adminJSON(ctx, http.MethodPatch, patchPath, map[string]string{
		"bucket":      bucket,
		"storageType": "primary",
	}, []int{http.StatusOK})
	if err == nil {
		return nil
	}
	if !isAdminNotFound(err) {
		return fmt.Errorf("assign primary storage %q: %w", storageID, err)
	}

	rootPath, maxBytes, err := m.templateLocalStorageConfig(ctx)
	if err != nil {
		return fmt.Errorf("resolve local storage template: %w", err)
	}
	createBody := map[string]any{
		"id":          storageID,
		"provider":    "local",
		"bucket":      bucket,
		"storageType": "primary",
		"label":       storageID,
		"enabled":     true,
		"config": map[string]string{
			"rootPath": rootPath,
			"maxBytes": maxBytes,
		},
	}
	if err := m.adminJSON(ctx, http.MethodPost, "/admin/api/storage/storages", createBody, []int{http.StatusCreated}); err != nil {
		return fmt.Errorf("create storage %q: %w", storageID, err)
	}
	return nil
}

func isAdminNotFound(err error) bool {
	return err != nil && (strings.Contains(err.Error(), " 404") || strings.Contains(strings.ToLower(err.Error()), "not found"))
}

func (m *miniCloudBackend) templateLocalStorageConfig(ctx context.Context) (rootPath, maxBytes string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.baseURL+"/admin/api/storage/storages", nil)
	if err != nil {
		return "", "", err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", "", fmt.Errorf("list storages: %s", strings.TrimSpace(string(b)))
	}
	var payload struct {
		Storages []struct {
			Provider string         `json:"provider"`
			Config   map[string]any `json:"config"`
		} `json:"storages"`
		Backends []struct {
			Provider string         `json:"provider"`
			Config   map[string]any `json:"config"`
		} `json:"backends"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", "", err
	}
	list := payload.Storages
	if len(list) == 0 {
		list = payload.Backends
	}
	for _, s := range list {
		if strings.EqualFold(s.Provider, "local") {
			if rp, ok := s.Config["rootPath"].(string); ok && strings.TrimSpace(rp) != "" {
				rootPath = strings.TrimSpace(rp)
			}
			if mb, ok := s.Config["maxBytes"].(string); ok && strings.TrimSpace(mb) != "" {
				maxBytes = strings.TrimSpace(mb)
			} else if mb, ok := s.Config["maxBytes"].(float64); ok && mb > 0 {
				maxBytes = fmt.Sprintf("%.0f", mb)
			}
			if rootPath != "" {
				if maxBytes == "" {
					maxBytes = strings.TrimSpace(config.MiniCloudLocalStorageQuota)
				}
				return rootPath, maxBytes, nil
			}
		}
	}
	return "", "", fmt.Errorf("no local storage backend found in MiniCloud to clone rootPath from")
}

func (m *miniCloudBackend) adminJSON(ctx context.Context, method, path string, body any, okStatuses []int) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, method, m.baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	for _, code := range okStatuses {
		if resp.StatusCode == code {
			return nil
		}
	}
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("%s %s: %d %s", method, path, resp.StatusCode, strings.TrimSpace(string(b)))
}

func (m *miniCloudBackend) putOnce(ctx context.Context, bucket, objectKey string, data []byte, contentType string) (string, error) {
	body, ct, err := buildGCSMultipartBody(objectKey, contentType, data)
	if err != nil {
		return "", err
	}
	uploadURL := fmt.Sprintf("%s/storage/upload/storage/v1/b/%s/o?uploadType=multipart",
		m.baseURL, url.PathEscape(bucket))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, uploadURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", ct)
	req.ContentLength = int64(len(body))
	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		return "", fmt.Errorf("minicloud upload failed: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	return FormatMC(bucket, objectKey), nil
}
