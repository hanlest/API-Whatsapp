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
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
)

type miniCloudBackend struct {
	client  *http.Client
	baseURL string
}

func newMiniCloudBackend() *miniCloudBackend {
	return &miniCloudBackend{
		client:  &http.Client{Timeout: 10 * time.Minute},
		baseURL: strings.TrimRight(strings.TrimSpace(config.MiniCloudBaseURL), "/"),
	}
}

func (m *miniCloudBackend) Put(ctx context.Context, deviceID, bucket, objectKey string, data []byte, contentType string) (string, error) {
	if m.baseURL == "" {
		return "", fmt.Errorf("minicloud base url is not configured")
	}
	if strings.TrimSpace(bucket) == "" {
		return "", fmt.Errorf("bucket is required")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	ref, err := m.putOnce(ctx, bucket, objectKey, data, contentType)
	if err == nil {
		return ref, nil
	}
	if isBucketNotConfiguredError(err) {
		if ensureErr := ensureMiniCloudBucket(ctx, m, deviceID, bucket); ensureErr != nil {
			return "", fmt.Errorf("%v; %w", err, ensureErr)
		}
		return m.putOnce(ctx, bucket, objectKey, data, contentType)
	}
	return "", err
}

func (m *miniCloudBackend) Get(ctx context.Context, ref *Ref) ([]byte, string, error) {
	rc, ct, err := m.OpenReader(ctx, ref)
	if err != nil {
		return nil, "", err
	}
	defer rc.Close()
	data, err := io.ReadAll(rc)
	return data, ct, err
}

func (m *miniCloudBackend) OpenReader(ctx context.Context, ref *Ref) (io.ReadCloser, string, error) {
	if ref == nil || ref.Bucket == "" || ref.ObjectKey == "" {
		return nil, "", fmt.Errorf("invalid minicloud ref")
	}
	if m.baseURL == "" {
		return nil, "", fmt.Errorf("minicloud base url is not configured")
	}
	objectPath := encodeObjectPath(ref.ObjectKey)
	mediaURL := fmt.Sprintf("%s/storage/storage/v1/b/%s/o/%s?alt=media",
		m.baseURL, url.PathEscape(ref.Bucket), objectPath)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mediaURL, nil)
	if err != nil {
		return nil, "", err
	}
	resp, err := m.client.Do(req)
	if err != nil {
		return nil, "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()
		return nil, "", fmt.Errorf("minicloud get failed: %s: %s", resp.Status, strings.TrimSpace(string(b)))
	}
	ct := resp.Header.Get("Content-Type")
	return resp.Body, ct, nil
}

func encodeObjectPath(objectKey string) string {
	parts := strings.Split(strings.Trim(objectKey, "/"), "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return strings.Join(parts, "/")
}

func buildGCSMultipartBody(objectName, contentType string, data []byte) ([]byte, string, error) {
	boundary := "GowaMiniCloudBoundary"
	meta, err := json.Marshal(map[string]string{
		"name":        objectName,
		"contentType": contentType,
	})
	if err != nil {
		return nil, "", err
	}
	var buf bytes.Buffer
	write := func(s string) { buf.WriteString(s) }
	write("--" + boundary + "\r\n")
	write("Content-Type: application/json; charset=UTF-8\r\n\r\n")
	write(string(meta))
	write("\r\n")
	write("--" + boundary + "\r\n")
	write("Content-Type: " + contentType + "\r\n\r\n")
	buf.Write(data)
	write("\r\n")
	write("--" + boundary + "--\r\n")
	return buf.Bytes(), "multipart/related; boundary=" + boundary, nil
}
