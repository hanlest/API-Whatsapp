package send

// ButtonsRequest matches Exegeta/Baileys POST /whatsapp/send-buttons (GoWA renders options as numbered plain text).
type ButtonsRequest struct {
	Phone   string         `json:"phone"`
	Text    string         `json:"text"`
	Footer  *string        `json:"footer"`
	Buttons []ButtonOption `json:"buttons"`
}

type ButtonOption struct {
	ID          string `json:"id"`
	DisplayText string `json:"displayText"`
}
