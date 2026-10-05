package send

import "mime/multipart"

// StatusRequest publishes a WhatsApp Status (24h story) to status@broadcast.
// JSON shape matches Exegeta/Baileys POST /whatsapp/send-status for drop-in use.
type StatusRequest struct {
	Kind            string                `json:"kind" form:"kind"`
	Text            string                `json:"text,omitempty" form:"text"`
	BackgroundColor string                `json:"backgroundColor,omitempty" form:"backgroundColor"`
	Font            *int                  `json:"font,omitempty" form:"font"`
	FileBase64      string                `json:"fileBase64,omitempty" form:"fileBase64"`
	Mimetype        string                `json:"mimetype,omitempty" form:"mimetype"`
	Caption         string                `json:"caption,omitempty" form:"caption"`
	// StatusJids is optional. When empty, WhatsApp status privacy + contact list on the
	// linked device is used (like the official app). Exegeta/Baileys send an explicit list.
	StatusJids      []string              `json:"statusJids,omitempty" form:"statusJids"`
	Attachment      *multipart.FileHeader `json:"-" form:"attachment"`
}
