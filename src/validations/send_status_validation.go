package validations

import (
	"context"
	"fmt"
	"mime/multipart"
	"regexp"
	"strings"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
)

var statusBackgroundColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

const (
	statusTextMaxLen = 700
	statusFontMin    = 0
	statusFontMax    = 8
)

func ValidateSendStatus(ctx context.Context, request domainSend.StatusRequest) error {
	kind := strings.TrimSpace(strings.ToLower(request.Kind))
	if kind == "" {
		return pkgError.ValidationError("kind is required")
	}
	if kind != "text" && kind != "image" && kind != "video" {
		return pkgError.ValidationError("kind must be one of: text, image, video")
	}

	for i, jid := range request.StatusJids {
		jid = strings.TrimSpace(jid)
		if jid == "" {
			return pkgError.ValidationError(fmt.Sprintf("statusJids[%d] cannot be empty", i))
		}
		if !strings.Contains(jid, "@") {
			return pkgError.ValidationError(fmt.Sprintf("statusJids[%d] must be a full JID (e.g. 569...@s.whatsapp.net)", i))
		}
		if strings.HasSuffix(jid, "@g.us") || jid == "status@broadcast" {
			return pkgError.ValidationError(fmt.Sprintf("statusJids[%d] must be a user JID", i))
		}
	}

	switch kind {
	case "text":
		text := strings.TrimSpace(request.Text)
		if text == "" {
			return pkgError.ValidationError("text is required when kind is text")
		}
		if len([]rune(text)) > statusTextMaxLen {
			return pkgError.ValidationError(fmt.Sprintf("text must be at most %d characters", statusTextMaxLen))
		}
		if request.FileBase64 != "" || request.Mimetype != "" || request.Attachment != nil {
			return pkgError.ValidationError("fileBase64, mimetype, and attachment are not allowed when kind is text")
		}
		if request.BackgroundColor != "" && !statusBackgroundColorPattern.MatchString(request.BackgroundColor) {
			return pkgError.ValidationError("backgroundColor must match #RRGGBB")
		}
		if request.Font != nil {
			if *request.Font < statusFontMin || *request.Font > statusFontMax {
				return pkgError.ValidationError(fmt.Sprintf("font must be between %d and %d", statusFontMin, statusFontMax))
			}
		}
	case "image", "video":
		if strings.TrimSpace(request.Text) != "" {
			return pkgError.ValidationError("text is not allowed when kind is image or video")
		}
		if request.BackgroundColor != "" || request.Font != nil {
			return pkgError.ValidationError("backgroundColor and font are only allowed when kind is text")
		}
		hasBase64 := strings.TrimSpace(request.FileBase64) != ""
		hasFile := request.Attachment != nil
		if hasBase64 == hasFile {
			return pkgError.ValidationError("provide exactly one of fileBase64 (JSON) or attachment (multipart)")
		}
		if strings.TrimSpace(request.Mimetype) == "" {
			return pkgError.ValidationError("mimetype is required when kind is image or video")
		}
		if request.Attachment != nil {
			if err := validateStatusAttachmentSize(request.Attachment); err != nil {
				return err
			}
		}
	}

	return nil
}

func validateStatusAttachmentSize(file *multipart.FileHeader) error {
	if file == nil {
		return nil
	}
	if file.Size <= 0 {
		return pkgError.ValidationError("attachment cannot be empty")
	}
	return nil
}
