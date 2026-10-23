package helper

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

var validate = newValidator()

func newValidator() *validator.Validate {
	v := validator.New()

	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})

	// Validasi custom NIM (hanya angka)
	_ = v.RegisterValidation("nimformat", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		if val == "" { return true } // biarkan tag 'required' yang menangani
		_, err := strconv.Atoi(val)
		return err == nil
	})

	// Perbaikan Bug 6: Logika validasi password dibalik
	_ = v.RegisterValidation("strongpassword", func(fl validator.FieldLevel) bool {
		return passwordStrength(fl.Field().String()) == ""
	})

	return v
}

func ValidateStruct(s any) map[string]string {
	err := validate.Struct(s)
	if err == nil { return nil }

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return map[string]string{"_": "validasi gagal"}
	}

	result := make(map[string]string)
	for _, fe := range fieldErrors {
		if _, exists := result[fe.Field()]; !exists {
			result[fe.Field()] = messageFor(fe)
		}
	}
	return result
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required": return "wajib diisi"
	case "min":
		if fe.Kind() == reflect.String { return "minimal " + fe.Param() + " karakter" }
		return "nilai minimal " + fe.Param()
	case "max":
		if fe.Kind() == reflect.String { return "maksimal " + fe.Param() + " karakter" }
		return "nilai maksimal " + fe.Param()
	case "nimformat": return "format NIM harus berupa angka"
	case "strongpassword":
		if val, ok := fe.Value().(string); ok { return passwordStrength(val) }
		return "password tidak memenuhi syarat"
	case "oneof": return "harus salah satu dari: " + strings.ReplaceAll(fe.Param(), " ", ", ")
	default: return "tidak memenuhi aturan " + fe.Tag()
	}
}

func passwordStrength(password string) string {
	if len(password) < 8 {
		return "minimal 8 karakter"
	}

	var hasLetter, hasDigit bool
	for _, r := range password {
		switch {
		case unicode.IsLetter(r):
			hasLetter = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}
	if !hasLetter || !hasDigit {
		return "harus memuat huruf dan angka"
	}

	weak := map[string]bool{
		"password1":   true,
		"12345678":    true,
		"qwerty123":   true,
		"admin123":    true,
		"password123": true,
	}
	if weak[password] {
		return "password terlalu umum"
	}

	return ""
}