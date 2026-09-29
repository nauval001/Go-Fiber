package service

import (
	"strings"
	"unicode"
)

func CheckPasswordStrength(password string) string {
	if len(password) < 8 {
		return "minimal 8 karakter"
	}
	var hasLetter, hasDigit bool
	for _, r := range password {
		if unicode.IsLetter(r) { hasLetter = true }
		if unicode.IsDigit(r) { hasDigit = true }
	}
	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}
	
	weak := map[string]bool{
		"password123": true, "12345678": true, "admin123": true,
	}
	if weak[strings.ToLower(password)] {
		return "password terlalu umum"
	}
	return ""
}