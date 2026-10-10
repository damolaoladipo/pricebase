package unit

import (
	"strings"
	"testing"

	"github.com/damolaoladipo/pricebase/internal/dtos"
	"github.com/damolaoladipo/pricebase/internal/services"
)

func TestValidateRegister(t *testing.T) {
	svc := services.NewUserService(nil, nil, nil)
	valid := dtos.RegisterUserDTO{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     "ada@example.com",
		Password:  "MyS3cret!Pass",
		UserType:  "user",
	}

	if got := svc.ValidateRegister(valid); got.Error {
		t.Errorf("expected valid payload to pass: %+v", got)
	}

	cases := []struct {
		name   string
		mutate func(*dtos.RegisterUserDTO)
		want   string
	}{
		{"email", func(d *dtos.RegisterUserDTO) { d.Email = "" }, "Email is required"},
		{"first", func(d *dtos.RegisterUserDTO) { d.FirstName = "" }, "First name is required"},
		{"last", func(d *dtos.RegisterUserDTO) { d.LastName = "" }, "Last name is required"},
		{"password", func(d *dtos.RegisterUserDTO) { d.Password = "" }, "Password is required"},
		{"type", func(d *dtos.RegisterUserDTO) { d.UserType = "superadmin" }, "Invalid user type value"},
	}

	for _, tc := range cases {
		data := valid
		tc.mutate(&data)
		got := svc.ValidateRegister(data)
		if !got.Error {
			t.Errorf("%s: expected error", tc.name)
		}
		if !strings.Contains(got.Message, tc.want) {
			t.Errorf("%s: expected %q, got %q", tc.name, tc.want, got.Message)
		}
	}
}

func TestValidateLogin(t *testing.T) {
	svc := services.NewUserService(nil, nil, nil)

	if got := svc.ValidateLogin(dtos.LoginDTO{Email: "ada@example.com", Password: "x"}); got.Error {
		t.Errorf("expected valid login to pass: %+v", got)
	}
	if got := svc.ValidateLogin(dtos.LoginDTO{Password: "x"}); !got.Error || got.Message != "email is required" {
		t.Errorf("expected email-required, got %+v", got)
	}
	if got := svc.ValidateLogin(dtos.LoginDTO{Email: "a@b.c"}); !got.Error || got.Message != "password is required" {
		t.Errorf("expected password-required, got %+v", got)
	}
	if got := svc.ValidateLogin(dtos.LoginDTO{Email: "not-an-email", Password: "x"}); !got.Error {
		t.Errorf("expected bad email to fail: %+v", got)
	}
}

func TestValidatePhoneNumber(t *testing.T) {
	svc := services.NewUserService(nil, nil, nil)

	for _, phone := range []string{"08031234567", "09012345678"} {
		if !svc.ValidatePhoneNumber(phone) {
			t.Errorf("expected %q to validate", phone)
		}
	}
	for _, phone := range []string{"", "08", "18031234567", "0a031234567"} {
		if svc.ValidatePhoneNumber(phone) {
			t.Errorf("expected %q to fail", phone)
		}
	}
}

func TestCheckEmail(t *testing.T) {
	svc := services.NewUserService(nil, nil, nil)

	if !svc.CheckEmail("ada@example.com") {
		t.Error("expected valid email to pass")
	}
	if !svc.CheckEmail("anything@something.africa") {
		t.Error("expected .africa bypass")
	}
	if svc.CheckEmail("not-an-email") {
		t.Error("expected invalid email to fail")
	}
}

func TestCheckPassword(t *testing.T) {
	svc := services.NewUserService(nil, nil, nil)

	if !svc.CheckPassword("MyS3cret!Pass") {
		t.Error("expected strong password to pass")
	}
	for _, weak := range []string{"short1!", "alllowercase1!", "ALLUPPERCASE1!", "NoDigitsHere!", "NoSpecial123", "        "} {
		if svc.CheckPassword(weak) {
			t.Errorf("expected %q to fail", weak)
		}
	}
}

func TestValidateLoginCredentials(t *testing.T) {
	svc := services.NewUserService(nil, nil, nil)

	if got := svc.ValidateLoginCredentials(dtos.LoginDTO{Email: "a@b.c", Password: "x"}); got.Error {
		t.Errorf("expected pass: %+v", got)
	}
	if got := svc.ValidateLoginCredentials(dtos.LoginDTO{}); !got.Error {
		t.Error("expected empty credentials to fail")
	}
}

func TestValidateUserType(t *testing.T) {
	svc := services.NewUserService(nil, nil, nil)

	if !svc.ValidateUserType("user", []string{"user", "admin"}) {
		t.Error("expected listed type to pass")
	}
	if svc.ValidateUserType("superadmin", []string{"user", "admin"}) {
		t.Error("expected unlisted type to fail")
	}
	if svc.ValidateUserType("user", nil) {
		t.Error("expected empty allowlist to reject")
	}
}

func TestAttachPhoneCode(t *testing.T) {
	svc := services.NewUserService(nil, nil, nil)

	if got := svc.AttachPhoneCode("+234", "08031234567"); got != "+2348031234567" {
		t.Errorf("plus code mismatch: %q", got)
	}
	if got := svc.AttachPhoneCode("234-801", "08031234567"); got != "+-8018031234567" {
		t.Errorf("dashed code must mirror TS substring handling, got %q", got)
	}
	if got := svc.AttachPhoneCode("", "0803"); got != "" {
		t.Errorf("expected empty for missing code, got %q", got)
	}
	if got := svc.AttachPhoneCode("+234", ""); got != "" {
		t.Errorf("expected empty for missing phone, got %q", got)
	}
}

func TestCheckPhoneCode(t *testing.T) {
	svc := services.NewUserService(nil, nil, nil)

	if got := svc.CheckPhoneCode("+234", "08031234567"); got != "+23431234567" {
		t.Errorf("long local mismatch: %q", got)
	}
	if got := svc.CheckPhoneCode("+234", "+2348031234567"); got != "+2348031234567" {
		t.Errorf("prefixed number must pass through, got %q", got)
	}
	if got := svc.CheckPhoneCode("", "0803"); got != "" {
		t.Errorf("expected empty for missing code, got %q", got)
	}
}
