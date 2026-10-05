package validations

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var chileanBaileysPhoneRegex = regexp.MustCompile(`^569\d{8}$`)

func ValidateSendButtons(ctx context.Context, request domainSend.ButtonsRequest) error {
	err := validation.ValidateStructWithContext(ctx, &request,
		validation.Field(&request.Phone, validation.Required),
		validation.Field(&request.Text, validation.Required, validation.Length(1, 1024)),
		validation.Field(&request.Buttons, validation.Required, validation.Length(1, 3)),
	)
	if err != nil {
		return pkgError.ValidationError(err.Error())
	}

	phone := strings.TrimSpace(request.Phone)
	if !chileanBaileysPhoneRegex.MatchString(phone) {
		return pkgError.ValidationError(`"phone" must match pattern ^569\d{8}$`)
	}

	text := strings.TrimSpace(request.Text)
	if text == "" || utf8.RuneCountInString(text) > 1024 {
		return pkgError.ValidationError(`"text" length must be between 1 and 1024`)
	}

	if footer := normalizeButtonsFooter(request.Footer); footer != "" {
		if utf8.RuneCountInString(footer) > 60 {
			return pkgError.ValidationError(`"footer" length must be less than or equal to 60 characters long`)
		}
	}

	for i, btn := range request.Buttons {
		id := strings.TrimSpace(btn.ID)
		display := strings.TrimSpace(btn.DisplayText)
		if id == "" {
			return pkgError.ValidationError(`"buttons[` + strconv.Itoa(i) + `].id" is required`)
		}
		if display == "" {
			return pkgError.ValidationError(`"buttons[` + strconv.Itoa(i) + `].displayText" is required`)
		}
		if utf8.RuneCountInString(id) > 256 {
			return pkgError.ValidationError(`"buttons[` + strconv.Itoa(i) + `].id" length must be between 1 and 256`)
		}
		if utf8.RuneCountInString(display) < 1 || utf8.RuneCountInString(display) > 20 {
			return pkgError.ValidationError(`"buttons[` + strconv.Itoa(i) + `].displayText" length must be between 1 and 20`)
		}
	}

	return nil
}

func normalizeButtonsFooter(footer *string) string {
	if footer == nil {
		return ""
	}
	return strings.TrimSpace(*footer)
}
