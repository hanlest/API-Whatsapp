package storage

import (
	"context"
	"testing"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/stretchr/testify/require"
)

func TestEnsureMiniCloudBucketForDeviceSkipsLocalDriver(t *testing.T) {
	prevDriver := config.MediaStorageDriver
	prevAuto := config.MiniCloudAutoCreateBucket
	prevURL := config.MiniCloudBaseURL
	prevBucket := config.MiniCloudDefaultBucket
	t.Cleanup(func() {
		config.MediaStorageDriver = prevDriver
		config.MiniCloudAutoCreateBucket = prevAuto
		config.MiniCloudBaseURL = prevURL
		config.MiniCloudDefaultBucket = prevBucket
	})

	config.MediaStorageDriver = "local"
	config.MiniCloudAutoCreateBucket = true
	config.MiniCloudBaseURL = "http://localhost:8082"
	config.MiniCloudDefaultBucket = "b1"

	err := EnsureMiniCloudBucketForDevice(context.Background(), "dev-1")
	require.NoError(t, err)
}
