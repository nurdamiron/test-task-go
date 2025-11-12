package service

import (
	"errors"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	phoneRegex    = regexp.MustCompile(`^\+[1-9]\d{1,14}$`)
	fullNameRegex = regexp.MustCompile(`^[a-zA-Zа-яА-ЯёЁәіңғүұқөһӘІҢҒҮҰҚӨҺ\s\-]+$`)
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

func (v *ValidationErrors) Error() string {
	return "ошибка валидации"
}

func (v *ValidationErrors) Add(field, message string) {
	v.Errors = append(v.Errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

func (v *ValidationErrors) HasErrors() bool {
	return len(v.Errors) > 0
}

func ValidateFullName(fullName string) error {
	trimmed := strings.TrimSpace(fullName)
	length := utf8.RuneCountInString(trimmed)

	if length < 2 {
		return errors.New("минимум 2 символа")
	}
	if length > 200 {
		return errors.New("максимум 200 символов")
	}
	if !fullNameRegex.MatchString(trimmed) {
		return errors.New("только буквы, пробелы и дефисы")
	}
	return nil
}

func ValidatePhone(phone string) error {
	trimmed := strings.TrimSpace(phone)

	if !phoneRegex.MatchString(trimmed) {
		return errors.New("формат E.164 (+[1-15 цифр])")
	}
	return nil
}

func ValidateCity(city string) error {
	trimmed := strings.TrimSpace(city)
	length := utf8.RuneCountInString(trimmed)

	if length < 2 {
		return errors.New("минимум 2 символа")
	}
	if length > 120 {
		return errors.New("максимум 120 символов")
	}
	return nil
}

func NormalizeString(s string) string {
	return strings.TrimSpace(s)
}
