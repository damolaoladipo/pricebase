package unit

import (
	"encoding/hex"
	"testing"

	"github.com/damolaoladipo/pricebase/internal/utils"
)

func TestGenerateIDShape(t *testing.T) {
	id := utils.GenerateID()

	if len(id) != 24 {
		t.Fatalf("expected 24-char id, got %q (len %d)", id, len(id))
	}

	if _, err := hex.DecodeString(id); err != nil {
		t.Fatalf("expected hex id, got %q: %v", id, err)
	}
}

func TestGenerateIDUnique(t *testing.T) {
	seen := map[string]struct{}{}

	for range 1000 {
		id := utils.GenerateID()
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate id generated: %q", id)
		}
		seen[id] = struct{}{}
	}
}
