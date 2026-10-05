package validations

import (
	"context"
	"strings"
	"testing"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/stretchr/testify/assert"
)

func TestValidateSendButtons(t *testing.T) {
	ctx := context.Background()
	validBtn := domainSend.ButtonOption{ID: "opt1", DisplayText: "Sí"}

	tests := []struct {
		name    string
		req     domainSend.ButtonsRequest
		wantErr bool
		contains string
	}{
		{
			name: "valid minimal",
			req: domainSend.ButtonsRequest{
				Phone:   "56912345678",
				Text:    "Elige una opción",
				Buttons: []domainSend.ButtonOption{validBtn},
			},
		},
		{
			name: "invalid phone",
			req: domainSend.ButtonsRequest{
				Phone:   "5691234567",
				Text:    "hola",
				Buttons: []domainSend.ButtonOption{validBtn},
			},
			wantErr:  true,
			contains: "phone",
		},
		{
			name: "too many buttons",
			req: domainSend.ButtonsRequest{
				Phone: "56912345678",
				Text:  "hola",
				Buttons: []domainSend.ButtonOption{
					{ID: "1", DisplayText: "A"},
					{ID: "2", DisplayText: "B"},
					{ID: "3", DisplayText: "C"},
					{ID: "4", DisplayText: "D"},
				},
			},
			wantErr: true,
		},
		{
			name: "displayText too long",
			req: domainSend.ButtonsRequest{
				Phone:   "56912345678",
				Text:    "hola",
				Buttons: []domainSend.ButtonOption{{ID: "1", DisplayText: strings.Repeat("x", 21)}},
			},
			wantErr:  true,
			contains: "displayText",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSendButtons(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.contains != "" {
					assert.Contains(t, strings.ToLower(err.Error()), strings.ToLower(tt.contains))
				}
				var ve pkgError.ValidationError
				assert.ErrorAs(t, err, &ve)
				return
			}
			assert.NoError(t, err)
		})
	}
}
