package entities

import (
	"time"
)

type CustomerType string

const (
	CustomerTypeIndividual CustomerType = "individual"
	CustomerTypeCompany    CustomerType = "company"
)

type Customer struct {
	ID           int          `json:"id" db:"id"`
	CustomerCode string       `json:"customer_code" db:"customer_code"`
	FullName     string       `json:"full_name" db:"full_name"`
	Email        string       `json:"email" db:"email"`
	Phone        string       `json:"phone" db:"phone"`
	Address      string       `json:"address" db:"address"`
	City         string       `json:"city" db:"city"`
	IDNumber     string       `json:"id_number" db:"id_number"`
	CustomerType CustomerType `json:"customer_type" db:"customer_type"`
	CompanyName  *string      `json:"company_name" db:"company_name"`
	TaxNumber    string       `json:"tax_number" db:"tax_number"`
	Notes        string       `json:"notes" db:"notes"`
	IsActive     bool         `json:"is_active" db:"is_active"`
	CreatedAt    time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time   `json:"deleted_at,omitempty" db:"deleted_at"`
}

type CreateCustomerRequest struct {
	FullName     string       `json:"full_name" binding:"required,min=2,max=100"`
	Email        string       `json:"email" binding:"omitempty,email"`
	Phone        string       `json:"phone" binding:"required"`
	Address      string       `json:"address" binding:"omitempty"`
	City         string       `json:"city" binding:"omitempty"`
	IDNumber     string       `json:"id_number" binding:"omitempty"`
	CustomerType CustomerType `json:"customer_type" binding:"required,oneof=individual company"`
	CompanyName  string       `json:"company_name" binding:"omitempty"`
	TaxNumber    string       `json:"tax_number" binding:"omitempty"`
	Notes        string       `json:"notes" binding:"omitempty"`
}

type UpdateCustomerRequest struct {
	FullName     string       `json:"full_name" binding:"omitempty,min=2,max=100"`
	Email        string       `json:"email" binding:"omitempty,email"`
	Phone        string       `json:"phone" binding:"omitempty"`
	Address      string       `json:"address" binding:"omitempty"`
	City         string       `json:"city" binding:"omitempty"`
	IDNumber     string       `json:"id_number" binding:"omitempty"`
	CustomerType CustomerType `json:"customer_type" binding:"omitempty,oneof=individual company"`
	CompanyName  string       `json:"company_name" binding:"omitempty"`
	TaxNumber    string       `json:"tax_number" binding:"omitempty"`
	Notes        string       `json:"notes" binding:"omitempty"`
	IsActive     *bool        `json:"is_active" binding:"omitempty"`
}

type CustomerSearchRequest struct {
	Query        string       `json:"query" form:"query"`
	CustomerType CustomerType `json:"customer_type" form:"customer_type"`
	City         string       `json:"city" form:"city"`
	IsActive     *bool        `json:"is_active" form:"is_active"`
	Page         int          `json:"page" form:"page"`
	Limit        int          `json:"limit" form:"limit"`
	SortBy       string       `json:"sort_by" form:"sort_by"`
	SortOrder    string       `json:"sort_order" form:"sort_order"`
}
