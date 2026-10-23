package service

import (
	"testing"

	"api-students/app/model"
)

func TestCheckPasswordStrength(t *testing.T) {
	cases := []struct {
		password string
		wantErr  bool
	}{
		{"rahasia123", false},   // valid
		{"abc", true},            // terlalu pendek
		{"abcdefgh", true},       // tidak ada angka
		{"12345678", true},       // tidak ada huruf
		{"password1", true},      // ada di daftar lemah
		{"Qwerty123", false},     // valid
	}
	for _, tc := range cases {
		got := checkPasswordStrength(tc.password)
		if (got != "") != tc.wantErr {
			t.Errorf("password %q: harap error=%v, dapat %q", tc.password, tc.wantErr, got)
		}
	}
}

func TestValidateRegister(t *testing.T) {
	// Valid
	errs := ValidateRegister(model.RegisterRequest{
		Username: "sari",
		Email:    "sari@example.com",
		Password: "rahasia123",
	})
	if len(errs) != 0 {
		t.Errorf("harus lolos, dapat: %v", errs)
	}

	// Email salah + password lemah
	errs = ValidateRegister(model.RegisterRequest{
		Username: "sari",
		Email:    "bukan-email",
		Password: "short",
	})
	if _, ok := errs["email"]; !ok {
		t.Error("email salah seharusnya ditolak")
	}
	if _, ok := errs["password"]; !ok {
		t.Error("password lemah seharusnya ditolak")
	}

	// Username terlalu pendek
	errs = ValidateRegister(model.RegisterRequest{
		Username: "ab",
		Email:    "ab@example.com",
		Password: "rahasia123",
	})
	if _, ok := errs["username"]; !ok {
		t.Error("username terlalu pendek seharusnya ditolak")
	}
}

func TestValidateLogin(t *testing.T) {
	// Lengkap
	errs := ValidateLogin(model.LoginRequest{Username: "sari", Password: "x"})
	if len(errs) != 0 {
		t.Errorf("harus lolos, dapat: %v", errs)
	}
	// Kosong
	errs = ValidateLogin(model.LoginRequest{})
	if len(errs) != 2 {
		t.Errorf("harap 2 error, dapat: %v", errs)
	}
}