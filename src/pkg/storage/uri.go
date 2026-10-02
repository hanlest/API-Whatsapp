package storage

import (
	"fmt"
	"strings"
)

const mcPrefix = "mc:"

// Ref is a stored media reference (MiniCloud object or local path).
type Ref struct {
	Bucket    string
	ObjectKey string
	LocalPath string
}

// IsMediaRef reports whether path is a MiniCloud URI (mc:bucket/key).
func IsMediaRef(path string) bool {
	return strings.HasPrefix(path, mcPrefix)
}

// FormatMC builds mc:{bucket}/{objectKey}.
func FormatMC(bucket, objectKey string) string {
	bucket = strings.TrimSpace(bucket)
	objectKey = strings.Trim(strings.TrimSpace(objectKey), "/")
	if bucket == "" || objectKey == "" {
		return ""
	}
	return mcPrefix + bucket + "/" + objectKey
}

// ParseMC parses mc:{bucket}/{objectKey}. Returns nil if not a valid mc ref.
func ParseMC(path string) (*Ref, error) {
	if !IsMediaRef(path) {
		return nil, fmt.Errorf("not a minicloud media ref")
	}
	rest := strings.TrimPrefix(path, mcPrefix)
	slash := strings.Index(rest, "/")
	if slash <= 0 || slash >= len(rest)-1 {
		return nil, fmt.Errorf("invalid minicloud media ref")
	}
	return &Ref{
		Bucket:    rest[:slash],
		ObjectKey: rest[slash+1:],
	}, nil
}

// ObjectKeyFromProxyPath returns the object key suffix after device folder for proxy routes.
func ObjectKeyFromProxyPath(deviceFolder, wildcardPath string) string {
	wildcardPath = strings.Trim(strings.ReplaceAll(wildcardPath, "\\", "/"), "/")
	deviceFolder = strings.Trim(deviceFolder, "/")
	if deviceFolder == "" {
		return wildcardPath
	}
	prefix := deviceFolder + "/"
	if strings.HasPrefix(wildcardPath, prefix) {
		return wildcardPath
	}
	if wildcardPath == "" {
		return deviceFolder
	}
	return deviceFolder + "/" + wildcardPath
}
