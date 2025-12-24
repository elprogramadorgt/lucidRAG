package whatsapp

import "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/whatsapp/dto"

func toHookVerificationDTO(input string) dto.HookVerificationResponse {
	return dto.HookVerificationResponse{
		Challenge: input,
	}
}
