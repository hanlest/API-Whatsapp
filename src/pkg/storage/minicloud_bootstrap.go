package storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/sirupsen/logrus"
)

func miniCloudAutoProvisionEnabled() bool {
	driver := strings.ToLower(strings.TrimSpace(config.MediaStorageDriver))
	return driver == "minicloud" &&
		config.MiniCloudAutoCreateBucket &&
		strings.TrimSpace(config.MiniCloudBaseURL) != ""
}

// EnsureMiniCloudBucketForDevice registers the bucket (if needed) and binds the
// device storage primary in MiniCloud. No-op when driver is not minicloud or
// auto-create is disabled. deviceID is the device slot id from POST /devices.
func EnsureMiniCloudBucketForDevice(ctx context.Context, deviceID string) error {
	if !miniCloudAutoProvisionEnabled() {
		return nil
	}
	deviceID = strings.TrimSpace(deviceID)
	if deviceID == "" {
		return fmt.Errorf("device id is required")
	}
	target, err := ResolveForDevice(deviceID)
	if err != nil {
		return err
	}
	if target.Driver != "minicloud" {
		return nil
	}
	m := newMiniCloudBackend()
	return ensureMiniCloudBucket(ctx, m, deviceID, target.Bucket)
}

// EnsureMiniCloudBucketsForRegisteredDevices provisions MiniCloud for every device
// record in chat storage (startup). Failures are logged; other devices still run.
func EnsureMiniCloudBucketsForRegisteredDevices(ctx context.Context) {
	if !miniCloudAutoProvisionEnabled() {
		return
	}
	if repo == nil {
		logrus.Warn("[MEDIA] MiniCloud bootstrap skipped: chat storage repository not wired")
		return
	}
	records, err := repo.ListDeviceRecords()
	if err != nil {
		logrus.Warnf("[MEDIA] MiniCloud bootstrap: list devices: %v", err)
		return
	}
	if len(records) == 0 {
		logrus.Debug("[MEDIA] MiniCloud bootstrap: no registered devices")
		return
	}
	for _, rec := range records {
		if rec == nil || strings.TrimSpace(rec.DeviceID) == "" {
			continue
		}
		deviceID := rec.DeviceID
		if err := EnsureMiniCloudBucketForDevice(ctx, deviceID); err != nil {
			logrus.Warnf("[MEDIA] MiniCloud bootstrap device %q: %v", deviceID, err)
			continue
		}
		logrus.Infof("[MEDIA] MiniCloud bucket/storage ready for device %q", deviceID)
	}
}
