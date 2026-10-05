package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainSend "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/send"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/validations"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"google.golang.org/protobuf/proto"
)

func (service serviceSend) SendButtons(ctx context.Context, request domainSend.ButtonsRequest) (response domainSend.GenericResponse, err error) {
	if err = validations.ValidateSendButtons(ctx, request); err != nil {
		return response, err
	}

	client := whatsapp.ClientFromContext(ctx)
	if client == nil {
		return response, pkgError.ErrWaSessionNotReady
	}
	if !client.IsConnected() || !client.IsLoggedIn() {
		return response, pkgError.ErrWaSessionNotReady
	}

	phoneJID := strings.TrimSpace(request.Phone) + config.WhatsappTypeUser
	recipient, err := utils.ParseJID(phoneJID)
	if err != nil {
		return response, pkgError.InvalidJID(fmt.Sprintf("invalid phone JID: %v", err))
	}

	if config.WhatsappAccountValidation && !utils.IsOnWhatsapp(client, phoneJID) {
		return response, pkgError.InvalidJID(fmt.Sprintf("Phone %s is not on whatsapp", phoneJID))
	}

	content := whatsapp.FormatButtonsMenuText(request)
	msg := &waE2E.Message{
		ExtendedTextMessage: &waE2E.ExtendedTextMessage{
			Text:        proto.String(content),
			ContextInfo: &waE2E.ContextInfo{},
		},
	}

	ts, err := service.wrapSendMessage(ctx, client, recipient, msg, content)
	if err != nil {
		return response, err
	}

	response.MessageID = ts.ID
	response.Status = "Botones enviados correctamente"
	return response, nil
}
