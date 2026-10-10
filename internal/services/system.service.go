package services

import (
	"fmt"
	"strings"

	Encryption "github.com/damolaoladipo/common-go/crypto"
	commonutils "github.com/damolaoladipo/common-go/utils"
	"github.com/damolaoladipo/pricebase/internal/dtos"
)

type SystemService struct{}

func NewSystemService() *SystemService {
	return &SystemService{}
}

// EncryptData encrypts data.Payload with data.Password and returns
// "<hex-cipher><separator><hex-vector>".
//
// The payload is wrapped in a {"payload": ...} envelope before encryption
// so DecryptData can return the original value with its type preserved.
func (s *SystemService) EncryptData(data dtos.EncryptDataDTO) (string, error) {

	if data.Password == "" {
		return "", fmt.Errorf("password is required")
	}
	if data.Separator == "" {
		return "", fmt.Errorf("separator is required")
	}

	envelope, err := commonutils.ToJSON(map[string]any{"payload": data.Payload})
	if err != nil {
		return "", fmt.Errorf("failed to encode payload: %w", err)
	}

	encrypted := Encryption.EncryptAESGCM(string(envelope), data.Password)
	if encrypted.Error {
		return "", fmt.Errorf("encryption failed: %s", encrypted.Message)
	}

	cipher, ok := encrypted.Data.(string)
	if !ok {
		return "", fmt.Errorf("encryption failed: unexpected cipher format")
	}

	return cipher + data.Separator + encrypted.Vector, nil
}

// DecryptData reverses EncryptData and returns the inner payload.
func (s *SystemService) DecryptData(data dtos.DecryptDataDTO) (any, error) {

	if data.Password == "" {
		return nil, fmt.Errorf("password is required")
	}
	if data.Separator == "" {
		return nil, fmt.Errorf("separator is required")
	}

	combined, ok := data.Payload.(string)
	if !ok {
		return nil, fmt.Errorf("payload must be a string")
	}

	parts := strings.Split(combined, data.Separator)
	if len(parts) < 2 {
		return nil, fmt.Errorf("malformed payload")
	}

	decrypted := Encryption.DecryptAESGCM(parts[0], parts[1], data.Password)
	if decrypted.Error {
		return nil, fmt.Errorf("decryption failed: %s", decrypted.Message)
	}

	text, ok := decrypted.Data.(string)
	if !ok {
		return nil, fmt.Errorf("decryption failed: unexpected plain format")
	}

	var envelope struct {
		Payload any `json:"payload"`
	}

	if err := commonutils.FromJSON([]byte(text), &envelope); err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	return envelope.Payload, nil
}
