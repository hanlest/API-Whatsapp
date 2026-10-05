package usecase

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"strings"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/sirupsen/logrus"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/validations"
	"github.com/disintegration/imaging"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"google.golang.org/protobuf/proto"
)

func (service serviceSend) SendStatus(ctx context.Context, request domainSend.StatusRequest) (response domainSend.GenericResponse, err error) {
	if err = validations.ValidateSendStatus(ctx, request); err != nil {
		return response, err
	}

	client := whatsapp.ClientFromContext(ctx)
	if client == nil {
		return response, pkgError.ErrWaCLI
	}

	kind := strings.ToLower(strings.TrimSpace(request.Kind))
	var msg *waE2E.Message

	switch kind {
	case "text":
		msg, err = buildStatusTextMessage(request)
	case "image":
		msg, err = service.buildStatusImageMessage(ctx, client, request)
	case "video":
		msg, err = service.buildStatusVideoMessage(ctx, client, request)
	default:
		return response, pkgError.ValidationError("kind must be one of: text, image, video")
	}
	if err != nil {
		return response, err
	}

	ts, err := service.wrapSendStatusMessage(ctx, client, msg, request.StatusJids)
	if err != nil {
		return response, normalizeSendError(err)
	}

	response.MessageID = ts.ID
	response.Status = fmt.Sprintf("Estado publicado correctamente (timestamp: %s)", ts.Timestamp.Format(time.RFC3339))
	return response, nil
}

func buildStatusTextMessage(request domainSend.StatusRequest) (*waE2E.Message, error) {
	ext := &waE2E.ExtendedTextMessage{
		Text: proto.String(strings.TrimSpace(request.Text)),
	}
	if request.BackgroundColor != "" {
		argb, err := parseStatusBackgroundARGB(request.BackgroundColor)
		if err != nil {
			return nil, err
		}
		ext.BackgroundArgb = proto.Uint32(argb)
		ext.TextArgb = proto.Uint32(0xFFFFFFFF)
	}
	if request.Font != nil {
		font := waE2E.ExtendedTextMessage_FontType(*request.Font)
		ext.Font = &font
	}
	return &waE2E.Message{ExtendedTextMessage: ext}, nil
}

func parseStatusBackgroundARGB(hexColor string) (uint32, error) {
	hexColor = strings.TrimPrefix(strings.TrimSpace(hexColor), "#")
	if len(hexColor) != 6 {
		return 0, pkgError.ValidationError("backgroundColor must match #RRGGBB")
	}
	raw, err := hex.DecodeString(hexColor)
	if err != nil {
		return 0, pkgError.ValidationError("backgroundColor must match #RRGGBB")
	}
	// WhatsApp status text uses ARGB with full alpha.
	return 0xFF000000 | (uint32(raw[0]) << 16) | (uint32(raw[1]) << 8) | uint32(raw[2]), nil
}

func (service serviceSend) buildStatusImageMessage(ctx context.Context, client *whatsmeow.Client, request domainSend.StatusRequest) (*waE2E.Message, error) {
	data, mimetype, err := loadStatusMediaBytes(request, config.WhatsappSettingMaxImageSize)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(mimetype, "image/") {
		return nil, pkgError.ValidationError("mimetype must be an image type for kind image")
	}

	uploaded, err := service.uploadMedia(ctx, client, whatsmeow.MediaImage, data, types.StatusBroadcastJID)
	if err != nil {
		return nil, pkgError.InternalServerError(fmt.Sprintf("failed to upload image: %v", err))
	}

	thumb := statusJPEGThumbnail(data)
	width, height := uint32(0), uint32(0)
	if cfg, _, decErr := image.DecodeConfig(bytes.NewReader(data)); decErr == nil {
		width = uint32(cfg.Width)
		height = uint32(cfg.Height)
	}

	return &waE2E.Message{ImageMessage: &waE2E.ImageMessage{
		JPEGThumbnail: thumb,
		Caption:       proto.String(request.Caption),
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      proto.String(mimetype),
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(data))),
		Width:         proto.Uint32(width),
		Height:        proto.Uint32(height),
	}}, nil
}

func (service serviceSend) buildStatusVideoMessage(ctx context.Context, client *whatsmeow.Client, request domainSend.StatusRequest) (*waE2E.Message, error) {
	data, mimetype, err := loadStatusMediaBytes(request, config.WhatsappSettingMaxVideoSize)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(mimetype, "video/") {
		return nil, pkgError.ValidationError("mimetype must be a video type for kind video")
	}

	uploaded, err := service.uploadMedia(ctx, client, whatsmeow.MediaVideo, data, types.StatusBroadcastJID)
	if err != nil {
		return nil, pkgError.InternalServerError(fmt.Sprintf("failed to upload video: %v", err))
	}

	return &waE2E.Message{VideoMessage: &waE2E.VideoMessage{
		Caption:       proto.String(request.Caption),
		URL:           proto.String(uploaded.URL),
		DirectPath:    proto.String(uploaded.DirectPath),
		MediaKey:      uploaded.MediaKey,
		Mimetype:      proto.String(mimetype),
		FileEncSHA256: uploaded.FileEncSHA256,
		FileSHA256:    uploaded.FileSHA256,
		FileLength:    proto.Uint64(uint64(len(data))),
	}}, nil
}

func loadStatusMediaBytes(request domainSend.StatusRequest, maxSize int64) ([]byte, string, error) {
	mimetype := strings.TrimSpace(request.Mimetype)
	if request.Attachment != nil {
		data, err := utils.MultipartFileHeaderToBytes(request.Attachment)
		if err != nil {
			return nil, "", pkgError.ValidationError("failed to read attachment")
		}
		if int64(len(data)) > maxSize {
			return nil, "", pkgError.ValidationError(fmt.Sprintf("attachment exceeds maximum size of %d bytes", maxSize))
		}
		if ct := request.Attachment.Header.Get("Content-Type"); ct != "" && mimetype == "" {
			mimetype = ct
		}
		return data, mimetype, nil
	}

	raw := strings.TrimSpace(request.FileBase64)
	data, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, "", pkgError.ValidationError("fileBase64 must be valid base64")
	}
	if int64(len(data)) > maxSize {
		return nil, "", pkgError.ValidationError(fmt.Sprintf("fileBase64 exceeds maximum size of %d bytes", maxSize))
	}
	return data, mimetype, nil
}

func statusJPEGThumbnail(data []byte) []byte {
	src, err := imaging.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	resized := imaging.Resize(src, 72, 0, imaging.Lanczos)
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, resized, imaging.JPEG, imaging.JPEGQuality(80)); err != nil {
		return nil
	}
	return buf.Bytes()
}

func (service serviceSend) wrapSendStatusMessage(ctx context.Context, client *whatsmeow.Client, msg *waE2E.Message, statusJids []string) (whatsmeow.SendResponse, error) {
	ts, err := whatsapp.SendStatusMessage(ctx, client, msg, statusJids)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}

	senderJID := ""
	if client.Store.ID != nil {
		senderJID = client.Store.ID.String()
	}
	content := statusStoredContent(msg)
	go func() {
		storeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
		defer cancel()
		if err := service.chatStorageRepo.StoreSentMessageWithContext(storeCtx, ts.ID, senderJID, types.StatusBroadcastJID.String(), content, ts.Timestamp, msg); err != nil {
			logrus.Warnf("Failed to store sent status %s: %v", ts.ID, err)
		}
	}()
	return ts, nil
}

func statusStoredContent(msg *waE2E.Message) string {
	if msg == nil {
		return ""
	}
	if t := msg.GetExtendedTextMessage(); t != nil {
		return t.GetText()
	}
	if img := msg.GetImageMessage(); img != nil {
		if c := img.GetCaption(); c != "" {
			return c
		}
		return "🖼️ Status image"
	}
	if vid := msg.GetVideoMessage(); vid != nil {
		if c := vid.GetCaption(); c != "" {
			return c
		}
		return "🎥 Status video"
	}
	return "Status"
}
