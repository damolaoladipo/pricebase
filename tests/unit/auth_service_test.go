package unit

import (
	"reflect"
	"testing"

	"github.com/damolaoladipo/pricebase/internal/services"
	"github.com/damolaoladipo/pricebase/internal/utils"
)

// ExtractActions and ValidateActions are pure: exercise them with a nil
// service so no database is needed.
func TestExtractActions(t *testing.T) {
	svc := services.NewAuthService(nil, nil, nil, nil)

	got := svc.ExtractActions([]utils.PermissionAction{
		{Name: "create", Description: "Create users"},
		{Name: "read", Description: "Read users"},
	})

	want := []string{"create", "read"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}

	if got := svc.ExtractActions(nil); len(got) != 0 {
		t.Errorf("expected empty result for nil input, got %v", got)
	}
}

func TestValidateActions(t *testing.T) {
	svc := services.NewAuthService(nil, nil, nil, nil)
	available := []string{"create", "read", "update"}

	if !svc.ValidateActions(available, []string{"create", "read"}) {
		t.Error("expected subset to validate")
	}

	if svc.ValidateActions(available, []string{"create", "delete"}) {
		t.Error("expected unknown action to fail validation")
	}

	if !svc.ValidateActions(available, nil) {
		t.Error("expected empty request to validate")
	}
}
