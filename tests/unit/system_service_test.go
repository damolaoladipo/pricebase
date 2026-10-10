package unit

import (
	"testing"

	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/services"
)

func TestSystemServiceRoundTrip(t *testing.T) {
	svc := services.NewSystemService()

	encrypted, err := svc.EncryptData(dtos.EncryptDataDTO{
		Payload:   map[string]any{"role": "admin"},
		Password:  "test-password-123",
		Separator: "::",
	})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	decrypted, err := svc.DecryptData(dtos.DecryptDataDTO{
		Payload:   encrypted,
		Password:  "test-password-123",
		Separator: "::",
	})
	if err != nil {
		t.Fatalf("decrypt: %v", err)
	}

	payload, ok := decrypted.(map[string]any)
	if !ok || payload["role"] != "admin" {
		t.Errorf("round-trip mismatch, got %v", decrypted)
	}
}

func TestSystemServiceFailures(t *testing.T) {
	svc := services.NewSystemService()

	encrypted, err := svc.EncryptData(dtos.EncryptDataDTO{
		Payload:   "hello",
		Password:  "pw",
		Separator: "::",
	})
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	if _, err := svc.DecryptData(dtos.DecryptDataDTO{Payload: encrypted, Password: "wrong", Separator: "::"}); err == nil {
		t.Error("expected error for wrong password")
	}
	if _, err := svc.DecryptData(dtos.DecryptDataDTO{Payload: "no-separator", Password: "pw", Separator: "::"}); err == nil {
		t.Error("expected error for malformed payload")
	}
	if _, err := svc.EncryptData(dtos.EncryptDataDTO{Payload: "x", Password: "pw"}); err == nil {
		t.Error("expected error for missing separator")
	}
}
