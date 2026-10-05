package webhook

// EventOption describes a webhook event type for UI and API discovery.
type EventOption struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// EventCatalog returns all webhook event names GoWA can forward (see docs/webhook-payload.md).
func EventCatalog() []EventOption {
	return []EventOption{
		{Name: "message", Description: "Text, media, contact, location, poll, and other message types"},
		{Name: "message.reaction", Description: "Emoji reactions to messages"},
		{Name: "message.revoked", Description: "Deleted or revoked messages (delete for everyone)"},
		{Name: "message.edited", Description: "Edited messages"},
		{Name: "message.ack", Description: "Delivery and read receipts"},
		{Name: "message.deleted", Description: "Messages deleted for you"},
		{Name: "chat_presence", Description: "Typing and recording indicators"},
		{Name: "group.participants", Description: "Group member join, leave, promote, demote"},
		{Name: "group.joined", Description: "You were added to a group"},
		{Name: "label.edit", Description: "WhatsApp label metadata changed"},
		{Name: "label.association", Description: "Label applied to or removed from a chat"},
		{Name: "newsletter.joined", Description: "Subscribed to a newsletter or channel"},
		{Name: "newsletter.left", Description: "Unsubscribed from a newsletter"},
		{Name: "newsletter.message", Description: "New newsletter or channel post"},
		{Name: "newsletter.mute", Description: "Newsletter mute setting changed"},
		{Name: "call.offer", Description: "Incoming call"},
	}
}
