package whatsapp

// HookInput represents the webhook verification request from WhatsApp.
type HookInput struct {
	Mode        string `json:"hub.mode"`
	Challenge   string `json:"hub.challenge"`
	VerifyToken string `json:"hub.verify_token"`
}
