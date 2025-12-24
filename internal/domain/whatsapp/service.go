package whatsapp

type Service interface {
	VerifyWebhook(req HookInput, expectedToken string) (string, error)
}
