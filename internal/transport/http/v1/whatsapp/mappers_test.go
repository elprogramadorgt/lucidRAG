package whatsapp

import (
	"testing"

	"github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/whatsapp/dto"
)

func TestMapToHookInput(t *testing.T) {
	tests := []struct {
		name     string
		request  dto.HookRequest
		expected string
	}{
		{
			name: "maps all fields correctly",
			request: dto.HookRequest{
				Mode:        "subscribe",
				Challenge:   "challenge123",
				VerifyToken: "verify-token",
			},
			expected: "subscribe",
		},
		{
			name: "empty request",
			request: dto.HookRequest{
				Mode:        "",
				Challenge:   "",
				VerifyToken: "",
			},
			expected: "",
		},
		{
			name: "special characters",
			request: dto.HookRequest{
				Mode:        "subscribe",
				Challenge:   "challenge-with-special-chars_123!@#",
				VerifyToken: "token_with_underscores",
			},
			expected: "subscribe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapToHookInput(tt.request)

			if result.Mode != tt.request.Mode {
				t.Errorf("Mode: expected %q, got %q", tt.request.Mode, result.Mode)
			}
			if result.Challenge != tt.request.Challenge {
				t.Errorf("Challenge: expected %q, got %q", tt.request.Challenge, result.Challenge)
			}
			if result.VerifyToken != tt.request.VerifyToken {
				t.Errorf("VerifyToken: expected %q, got %q", tt.request.VerifyToken, result.VerifyToken)
			}
		})
	}
}

func TestToHookVerificationDTO(t *testing.T) {
	tests := []struct {
		name      string
		challenge string
	}{
		{
			name:      "simple challenge",
			challenge: "abc123",
		},
		{
			name:      "empty challenge",
			challenge: "",
		},
		{
			name:      "long challenge",
			challenge: "this-is-a-very-long-challenge-string-that-might-be-used-in-production",
		},
		{
			name:      "challenge with special chars",
			challenge: "challenge!@#$%^&*()_+-=[]{}|;':\",./<>?",
		},
		{
			name:      "unicode challenge",
			challenge: "challenge-你好-مرحبا-🎉",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toHookVerificationDTO(tt.challenge)

			if result.Challenge != tt.challenge {
				t.Errorf("Expected challenge %q, got %q", tt.challenge, result.Challenge)
			}
		})
	}
}
