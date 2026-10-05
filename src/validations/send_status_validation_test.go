package validations

import (
	"context"
	"testing"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/stretchr/testify/assert"
)

func TestValidateSendStatus(t *testing.T) {
	ctx := context.Background()
	font := 2

	tests := []struct {
		name    string
		req     domainSend.StatusRequest
		wantErr string
	}{
		{
			name: "text ok",
			req: domainSend.StatusRequest{
				Kind:       "text",
				Text:       "Hola",
				StatusJids: []string{"56912345678@s.whatsapp.net"},
			},
		},
		{
			name: "text without statusJids uses contact audience",
			req: domainSend.StatusRequest{
				Kind: "text",
				Text: "Hola",
			},
		},
		{
			name: "image missing media",
			req: domainSend.StatusRequest{
				Kind:       "image",
				Mimetype:   "image/jpeg",
				StatusJids: []string{"56912345678@s.whatsapp.net"},
			},
			wantErr: "fileBase64",
		},
		{
			name: "text with bad color",
			req: domainSend.StatusRequest{
				Kind:            "text",
				Text:            "Hola",
				BackgroundColor: "red",
				StatusJids:      []string{"56912345678@s.whatsapp.net"},
			},
			wantErr: "backgroundColor",
		},
		{
			name: "text font range",
			req: domainSend.StatusRequest{
				Kind:       "text",
				Text:       "Hola",
				Font:       &font,
				StatusJids: []string{"56912345678@s.whatsapp.net"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSendStatus(ctx, tt.req)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}
