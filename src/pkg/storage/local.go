package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
)

type localBackend struct{}

func (localBackend) Put(_ context.Context, _ string, _ string, objectKey string, data []byte, _ string) (string, error) {
	rel := filepath.FromSlash(objectKey)
	full := filepath.Join(config.PathMedia, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(full, data, 0o600); err != nil {
		return "", err
	}
	return full, nil
}

func (localBackend) Get(_ context.Context, ref *Ref) ([]byte, string, error) {
	path := refLocalPath(ref)
	if path == "" {
		return nil, "", fmt.Errorf("empty local path")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	return data, "", nil
}

func (localBackend) OpenReader(_ context.Context, ref *Ref) (io.ReadCloser, string, error) {
	path := refLocalPath(ref)
	if path == "" {
		return nil, "", fmt.Errorf("empty local path")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	return f, "", nil
}

func refLocalPath(ref *Ref) string {
	if ref == nil {
		return ""
	}
	if ref.LocalPath != "" {
		return ref.LocalPath
	}
	return ""
}
