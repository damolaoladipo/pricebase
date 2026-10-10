package unit

import (
	"testing"

	"github.com/damolaoladipo/pricebase/internal/utils"
)

// Each enum is checked against a known-good literal from its const block
// (not the constant itself, so the test isn't tautological), a known-bad
// value, and its expected member count.
func TestEnums(t *testing.T) {
	cases := []struct {
		name     string
		isValid  func(string) bool
		count    func() int
		good     string
		expected int
	}{
		{"UserType", func(s string) bool { return utils.UserType(s).IsValid() }, func() int { return len(utils.UserTypeValues()) }, "superadmin", 5},
		{"PasswordType", func(s string) bool { return utils.PasswordType(s).IsValid() }, func() int { return len(utils.PasswordTypeValues()) }, "self", 3},
		{"OTPType", func(s string) bool { return utils.OTPType(s).IsValid() }, func() int { return len(utils.OTPTypeValues()) }, "phone", 2},
		{"OTPDestination", func(s string) bool { return utils.OTPDestination(s).IsValid() }, func() int { return len(utils.OTPDestinationValues()) }, "email", 2},
		{"LoginMethod", func(s string) bool { return utils.LoginMethod(s).IsValid() }, func() int { return len(utils.LoginMethodValues()) }, "biometric", 4},
		{"AppChannel", func(s string) bool { return utils.AppChannel(s).IsValid() }, func() int { return len(utils.AppChannelValues()) }, "web", 4},
		{"Permission", func(s string) bool { return utils.Permission(s).IsValid() }, func() int { return len(utils.PermissionValues()) }, "create", 11},
		{"Resource", func(s string) bool { return utils.Resource(s).IsValid() }, func() int { return len(utils.ResourceValues()) }, "role", 18},
		{"BusinessType", func(s string) bool { return utils.BusinessType(s).IsValid() }, func() int { return len(utils.BusinessTypeValues()) }, "corporate", 3},
		{"Prefix", func(s string) bool { return utils.Prefix(s).IsValid() }, func() int { return len(utils.PrefixValues()) }, "BIZ", 8},
		{"Domain", func(s string) bool { return utils.Domain(s).IsValid() }, func() int { return len(utils.DomainValues()) }, "live", 4},
		{"APIKeyEnum", func(s string) bool { return utils.APIKeyEnum(s).IsValid() }, func() int { return len(utils.APIKeyEnumValues()) }, "platform", 3},
		{"LoginType", func(s string) bool { return utils.LoginType(s).IsValid() }, func() int { return len(utils.LoginTypeValues()) }, "email", 2},
		{"SMSProvider", func(s string) bool { return utils.SMSProvider(s).IsValid() }, func() int { return len(utils.SMSProviderValues()) }, "aws", 3},
		{"Currency", func(s string) bool { return utils.Currency(s).IsValid() }, func() int { return len(utils.CurrencyValues()) }, "NGN", 3},
		{"CurrencyName", func(s string) bool { return utils.CurrencyName(s).IsValid() }, func() int { return len(utils.CurrencyNameValues()) }, "naira", 2},
		{"Action", func(s string) bool { return utils.Action(s).IsValid() }, func() int { return len(utils.ActionValues()) }, "generate", 10},
		{"RegisterMethod", func(s string) bool { return utils.RegisterMethod(s).IsValid() }, func() int { return len(utils.RegisterMethodValues()) }, "email", 2},
		{"CACType", func(s string) bool { return utils.CACType(s).IsValid() }, func() int { return len(utils.CACTypeValues()) }, "business-name", 5},
		{"CACSubType", func(s string) bool { return utils.CACSubType(s).IsValid() }, func() int { return len(utils.CACSubTypeValues()) }, "private-company", 4},
		{"CACDocument", func(s string) bool { return utils.CACDocument(s).IsValid() }, func() int { return len(utils.CACDocumentValues()) }, "certificate", 13},
		{"OnboardStage", func(s string) bool { return utils.OnboardStage(s).IsValid() }, func() int { return len(utils.OnboardStageValues()) }, "business-information", 3},
		{"AIModel", func(s string) bool { return utils.AIModel(s).IsValid() }, func() int { return len(utils.AIModelValues()) }, "openai", 3},
		{"Status", func(s string) bool { return utils.Status(s).IsValid() }, func() int { return len(utils.StatusValues()) }, "pending", 24},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !tc.isValid(tc.good) {
				t.Errorf("expected %q to be valid", tc.good)
			}
			if tc.isValid("bogus-value") {
				t.Errorf("expected %q to be invalid", "bogus-value")
			}
			if tc.isValid("") {
				t.Errorf("expected empty string to be invalid")
			}
			if got := tc.count(); got != tc.expected {
				t.Errorf("expected %d members, got %d", tc.expected, got)
			}
		})
	}
}

func TestUserTypeString(t *testing.T) {
	if got := utils.UserTypeSuper.String(); got != "superadmin" {
		t.Errorf("expected %q, got %q", "superadmin", got)
	}
}
