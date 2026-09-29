package service

import "testing"

func TestCheckPasswordStrength_TooShort(t *testing.T) {
	if msg := CheckPasswordStrength("abc12"); msg != "minimal 8 karakter" {
		t.Errorf("Harap error minimal 8 karakter, dapat: %s", msg)
	}
}

func TestCheckPasswordStrength_NoDigit(t *testing.T) {
	if msg := CheckPasswordStrength("passwordkuat"); msg != "harus memuat huruf dan angka" {
		t.Errorf("Harap error huruf dan angka, dapat: %s", msg)
	}
}

func TestCheckPasswordStrength_WeakPassword(t *testing.T) {
	if msg := CheckPasswordStrength("password123"); msg != "password terlalu umum" {
		t.Errorf("Harap error password terlalu umum, dapat: %s", msg)
	}
}