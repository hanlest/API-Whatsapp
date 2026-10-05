package whatsapp

import (
	"fmt"
	"strings"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
)

const buttonsMenuPrompt = "Responde con el número de tu opción:"

// FormatButtonsMenuText turns Exegeta/Baileys sendButtons payload into a plain-text menu
// (native WhatsApp buttons often do not render on the recipient's phone via multidevice).
func FormatButtonsMenuText(request domainSend.ButtonsRequest) string {
	var b strings.Builder
	body := strings.TrimSpace(request.Text)
	if body != "" {
		b.WriteString(body)
		b.WriteString("\n\n")
	}
	b.WriteString(buttonsMenuPrompt)
	b.WriteString("\n")
	for i, opt := range request.Buttons {
		display := strings.TrimSpace(opt.DisplayText)
		id := strings.TrimSpace(opt.ID)
		line := fmt.Sprintf("%d. %s", i+1, display)
		if id != "" && !strings.EqualFold(id, display) {
			line += fmt.Sprintf(" (%s)", id)
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	if request.Footer != nil {
		if footer := strings.TrimSpace(*request.Footer); footer != "" {
			b.WriteString("\n")
			b.WriteString(footer)
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
