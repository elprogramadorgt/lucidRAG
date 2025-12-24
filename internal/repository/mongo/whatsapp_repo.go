package mongo

type WhatsappRepo struct {
	c *DbClient
}

func NewWhatsappRepo(c *DbClient) *WhatsappRepo {
	return &WhatsappRepo{c: c}
}
