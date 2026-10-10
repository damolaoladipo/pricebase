package unit

import (
	"testing"

	"github.com/damolaoladipo/pricebase/internal/utils"
)

func TestNewErrorResponse(t *testing.T) {
	err := utils.NewErrorResponse("boom", 500, nil, nil)
	if err.Message != "boom" || err.Status != 500 {
		t.Errorf("fields mismatch: %+v", err)
	}
	if err.Errors == nil || len(err.Errors) != 0 {
		t.Errorf("nil errors must default to empty, got %+v", err.Errors)
	}
	if err.Error() != "boom" {
		t.Errorf("Error() must return the message, got %q", err.Error())
	}
}

func TestErrorsAsAny(t *testing.T) {
	got := utils.ErrorsAsAny([]string{"a", "b"})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Errorf("conversion mismatch: %v", got)
	}
	if got := utils.ErrorsAsAny(nil); len(got) != 0 {
		t.Errorf("nil must convert to empty, got %v", got)
	}
}
