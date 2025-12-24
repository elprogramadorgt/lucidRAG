package whatsapp

type Service interface {
	VerifyWebhook(req HookInput) (string, error)
}
