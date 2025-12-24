package whatsapp

import (
	whatsappDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/whatsapp"
	"github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/whatsapp/dto"
)

func mapToHookInput(req dto.HookRequest) whatsappDomain.HookInput {
	return whatsappDomain.HookInput{
		Mode:        req.Mode,
		Challenge:   req.Challenge,
		VerifyToken: req.VerifyToken,
	}
}
