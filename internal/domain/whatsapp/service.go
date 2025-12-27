package whatsapp

// Service defines the business operations for WhatsApp integration.
type Service interface {
	VerifyWebhook(req HookInput, expectedToken string) (string, error)
}
