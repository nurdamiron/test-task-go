package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFullName(t *testing.T) {
	tests := []struct {
		name      string
		fullName  string
		wantError bool
	}{
		{
			name:      "валидное имя латиницей",
			fullName:  "John Doe",
			wantError: false,
		},
		{
			name:      "валидное имя кириллицей",
			fullName:  "Иван Иванов",
			wantError: false,
		},
		{
			name:      "имя с дефисом",
			fullName:  "Анна-Мария",
			wantError: false,
		},
		{
			name:      "казахское имя",
			fullName:  "Әлихан Нұрғалиев",
			wantError: false,
		},
		{
			name:      "казахский город",
			fullName:  "Қарағанды",
			wantError: false,
		},
		{
			name:      "слишком короткое",
			fullName:  "A",
			wantError: true,
		},
		{
			name:      "пустая строка",
			fullName:  "",
			wantError: true,
		},
		{
			name:      "содержит цифры",
			fullName:  "John123",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFullName(tt.fullName)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidatePhone(t *testing.T) {
	tests := []struct {
		name      string
		phone     string
		wantError bool
	}{
		{
			name:      "валидный E.164",
			phone:     "+79991234567",
			wantError: false,
		},
		{
			name:      "валидный международный",
			phone:     "+12025551234",
			wantError: false,
		},
		{
			name:      "без плюса",
			phone:     "79991234567",
			wantError: true,
		},
		{
			name:      "начинается с нуля",
			phone:     "+09991234567",
			wantError: true,
		},
		{
			name:      "слишком короткий",
			phone:     "+1",
			wantError: true,
		},
		{
			name:      "пустая строка",
			phone:     "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePhone(tt.phone)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCity(t *testing.T) {
	tests := []struct {
		name      string
		city      string
		wantError bool
	}{
		{
			name:      "валидный город",
			city:      "Москва",
			wantError: false,
		},
		{
			name:      "длинное название",
			city:      "Санкт-Петербург",
			wantError: false,
		},
		{
			name:      "казахский город",
			city:      "Алматы",
			wantError: false,
		},
		{
			name:      "казахский город с қ",
			city:      "Қарағанды",
			wantError: false,
		},
		{
			name:      "слишком короткое",
			city:      "A",
			wantError: true,
		},
		{
			name:      "пустая строка",
			city:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCity(tt.city)
			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
