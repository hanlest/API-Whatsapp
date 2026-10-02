package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MediaKind is the top-level folder under the device folder.
type MediaKind string

const (
	MediaKindIncoming MediaKind = "incoming"
	MediaKindOutgoing MediaKind = "outgoing"
)

// BuildObjectKey builds {deviceFolder}/{kind}/{date}/{timestamp}-{uuid}{ext}.
func BuildObjectKey(deviceFolder string, kind MediaKind, ext string) string {
	deviceFolder = strings.Trim(deviceFolder, "/")
	if kind == "" {
		kind = MediaKindIncoming
	}
	date := time.Now().Format("2006-01-02")
	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return fmt.Sprintf("%s/%s/%s/%d-%s%s", deviceFolder, kind, date, time.Now().Unix(), uuid.NewString(), ext)
}
