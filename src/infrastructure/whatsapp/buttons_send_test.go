package whatsapp

import (
	"strings"
	"testing"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/stretchr/testify/assert"
)

func TestFormatButtonsMenuText(t *testing.T) {
	footer := "Pie de página"
	req := domainSend.ButtonsRequest{
		Text:   "Elige una opción",
		Footer: &footer,
		Buttons: []domainSend.ButtonOption{
			{ID: "ok", DisplayText: "Entendido"},
			{ID: "more", DisplayText: "Más info"},
		},
	}

	got := FormatButtonsMenuText(req)
	assert.Contains(t, got, "Elige una opción")
	assert.Contains(t, got, buttonsMenuPrompt)
	assert.Contains(t, got, "1. Entendido (ok)")
	assert.Contains(t, got, "2. Más info (more)")
	assert.True(t, strings.HasSuffix(got, "Pie de página"))
}

func TestFormatButtonsMenuTextOmitsRedundantID(t *testing.T) {
	req := domainSend.ButtonsRequest{
		Text: "Hola",
		Buttons: []domainSend.ButtonOption{
			{ID: "si", DisplayText: "si"},
		},
	}
	got := FormatButtonsMenuText(req)
	assert.Contains(t, got, "1. si")
	assert.NotContains(t, got, "(si)")
}
