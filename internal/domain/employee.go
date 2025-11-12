package domain

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID        uuid.UUID `json:"id"`
	FullName  string    `json:"fullName"`
	Phone     string    `json:"phone"`
	City      string    `json:"city"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type CreateEmployeeRequest struct {
	FullName string `json:"fullName"`
	Phone    string `json:"phone"`
	City     string `json:"city"`
}
