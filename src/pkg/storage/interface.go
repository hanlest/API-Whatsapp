package storage

import (
	"context"
	"io"
)

// Backend stores and retrieves media bytes.
type Backend interface {
	Put(ctx context.Context, deviceID, bucket, objectKey string, data []byte, contentType string) (ref string, err error)
	Get(ctx context.Context, ref *Ref) ([]byte, string, error)
	OpenReader(ctx context.Context, ref *Ref) (io.ReadCloser, string, error)
}
