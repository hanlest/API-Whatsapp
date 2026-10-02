package cmd

import (
	"context"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/storage"
)

// startMiniCloudMediaBootstrap ensures MiniCloud buckets/storages for registered devices.
func startMiniCloudMediaBootstrap() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
		defer cancel()
		storage.EnsureMiniCloudBucketsForRegisteredDevices(ctx)
	}()
}
