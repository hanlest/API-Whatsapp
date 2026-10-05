package whatsapp

import (
	"context"
	"fmt"
	"strings"
	"time"

	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
)

// ParseStatusRecipientJIDs parses Exegeta-style statusJids into whatsmeow JIDs.
func ParseStatusRecipientJIDs(statusJids []string) ([]types.JID, error) {
	out := make([]types.JID, 0, len(statusJids))
	seen := make(map[string]struct{}, len(statusJids))
	for i, raw := range statusJids {
		raw = strings.TrimSpace(raw)
		parsed, err := utils.ParseJID(raw)
		if err != nil {
			return nil, pkgError.ValidationError(fmt.Sprintf("statusJids[%d]: %v", i, err))
		}
		key := parsed.ToNonAD().String()
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, parsed.ToNonAD())
	}
	if len(out) == 0 {
		return nil, pkgError.ValidationError("statusJids must contain at least one valid JID")
	}
	return out, nil
}

// StatusSendTimeout estimates how long to wait for WhatsApp to ack a status publish.
func StatusSendTimeout(recipientCount int) time.Duration {
	const (
		base       = 2 * time.Minute
		perContact = 250 * time.Millisecond
		maxWait    = 15 * time.Minute
	)
	wait := base + time.Duration(recipientCount)*perContact
	if wait > maxWait {
		return maxWait
	}
	return wait
}

// SendStatusMessage publishes to status@broadcast.
// Empty statusJids uses WhatsApp status privacy + contacts (official-app behavior).
// Non-empty statusJids uses an explicit audience (Exegeta/Baileys statusJidList).
func SendStatusMessage(ctx context.Context, client *whatsmeow.Client, message *waE2E.Message, statusJids []string) (whatsmeow.SendResponse, error) {
	utils.MustLogin(client)

	to := types.StatusBroadcastJID
	extra := whatsmeow.SendRequestExtra{
		Timeout: 15 * time.Minute,
	}

	if len(statusJids) > 0 {
		participants, err := ParseStatusRecipientJIDs(statusJids)
		if err != nil {
			return whatsmeow.SendResponse{}, err
		}
		extra.StatusBroadcastParticipants = participants
		extra.Timeout = StatusSendTimeout(len(participants))
	}

	ts, err := client.SendMessage(ctx, to, message, extra)
	if err != nil {
		return whatsmeow.SendResponse{}, err
	}
	return ts, nil
}
