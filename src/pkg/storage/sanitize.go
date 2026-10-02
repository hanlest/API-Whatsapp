package storage

import (
	"regexp"
	"strings"
)

var unsafeFolderChars = regexp.MustCompile(`[/\\%#?*\x00-\x1f]+`)

// SanitizeDeviceFolder returns a safe folder name for object keys.
func SanitizeDeviceFolder(deviceID, folderOverride string) string {
	name := strings.TrimSpace(folderOverride)
	if name == "" {
		name = strings.TrimSpace(deviceID)
	}
	name = strings.ReplaceAll(name, " ", "-")
	name = unsafeFolderChars.ReplaceAllString(name, "")
	name = strings.Trim(name, ".")
	if name == "" {
		return "device"
	}
	return name
}
