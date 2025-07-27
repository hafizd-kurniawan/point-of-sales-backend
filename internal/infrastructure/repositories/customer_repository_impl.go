package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"

	"github.com/jmoiron/sqlx"
)

type customerRepositoryImpl struct {
	db *sqlx.DB
}

func NewCustomerRepository(db *sqlx.DB) repositories.CustomerRepository {
	return &customerRepositoryImpl{db: db}
}

func (r *customerRepositoryImpl) Create(customer *entities.Customer) error {
	// Generate customer code if not provided
	if customer.CustomerCode == "" {
		code, err := r.GenerateCustomerCode()
		if err != nil {
			return fmt.Errorf("failed to generate customer code: %w", err)
		}
		customer.CustomerCode = code
	}

	query := `
		INSERT INTO customers (customer_code, full_name, email, phone, address, city, 
		                      id_number, customer_type, company_name, tax_number, notes, is_active)
		VALUES (:customer_code, :full_name, :email, :phone, :address, :city, 
		        :id_number, :customer_type, :company_name, :tax_number, :notes, :is_active)
		RETURNING id, created_at, updated_at
	`

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	err = stmt.Get(customer, customer)
	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

func (r *customerRepositoryImpl) GetByID(id int) (*entities.Customer, error) {
	query := `
		SELECT id, customer_code, full_name, email, phone, address, city, id_number,
		       customer_type, company_name, tax_number, notes, is_active,
		       created_at, updated_at, deleted_at
		FROM customers 
		WHERE id = $1 AND deleted_at IS NULL
	`

	var customer entities.Customer
	err := r.db.Get(&customer, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

func (r *customerRepositoryImpl) GetByCustomerCode(customerCode string) (*entities.Customer, error) {
	query := `
		SELECT id, customer_code, full_name, email, phone, address, city, id_number,
		       customer_type, company_name, tax_number, notes, is_active,
		       created_at, updated_at, deleted_at
		FROM customers 
		WHERE customer_code = $1 AND deleted_at IS NULL
	`

	var customer entities.Customer
	err := r.db.Get(&customer, query, customerCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("customer not found")
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return &customer, nil
}

func (r *customerRepositoryImpl) GetAll(limit, offset int) ([]entities.Customer, error) {
	query := `
		SELECT id, customer_code, full_name, email, phone, address, city, id_number,
		       customer_type, company_name, tax_number, notes, is_active,
		       created_at, updated_at, deleted_at
		FROM customers 
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var customers []entities.Customer
	err := r.db.Select(&customers, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get customers: %w", err)
	}

	return customers, nil
}

func (r *customerRepositoryImpl) Search(req *entities.CustomerSearchRequest) ([]entities.Customer, int, error) {
	var whereConditions []string
	var args []interface{}
	argIndex := 1

	// Base where condition
	whereConditions = append(whereConditions, "deleted_at IS NULL")

	// Search query
	if req.Query != "" {
		whereConditions = append(whereConditions,
			fmt.Sprintf("(full_name ILIKE $%d OR email ILIKE $%d OR phone ILIKE $%d OR customer_code ILIKE $%d)",
				argIndex, argIndex, argIndex, argIndex))
		args = append(args, "%"+req.Query+"%")
		argIndex++
	}

	// Customer type filter
	if req.CustomerType != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("customer_type = $%d", argIndex))
		args = append(args, req.CustomerType)
		argIndex++
	}

	// City filter
	if req.City != "" {
		whereConditions = append(whereConditions, fmt.Sprintf("city ILIKE $%d", argIndex))
		args = append(args, "%"+req.City+"%")
		argIndex++
	}

	// Active status filter
	if req.IsActive != nil {
		whereConditions = append(whereConditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, *req.IsActive)
		argIndex++
	}

	whereClause := strings.Join(whereConditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM customers WHERE %s", whereClause)
	var total int
	err := r.db.Get(&total, countQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	// Set defaults
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 10
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortOrder == "" {
		req.SortOrder = "DESC"
	}

	offset := (req.Page - 1) * req.Limit

	// Build order clause
	orderClause := fmt.Sprintf("ORDER BY %s %s", req.SortBy, req.SortOrder)

	// Search query
	searchQuery := fmt.Sprintf(`
		SELECT id, customer_code, full_name, email, phone, address, city, id_number,
		       customer_type, company_name, tax_number, notes, is_active,
		       created_at, updated_at, deleted_at
		FROM customers 
		WHERE %s
		%s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderClause, argIndex, argIndex+1)

	args = append(args, req.Limit, offset)

	var customers []entities.Customer
	err = r.db.Select(&customers, searchQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search customers: %w", err)
	}

	return customers, total, nil
}

func (r *customerRepositoryImpl) Update(id int, customer *entities.Customer) error {
	query := `
		UPDATE customers 
		SET full_name = COALESCE(NULLIF(:full_name, ''), full_name),
		    email = COALESCE(NULLIF(:email, ''), email),
		    phone = COALESCE(NULLIF(:phone, ''), phone),
		    address = COALESCE(NULLIF(:address, ''), address),
		    city = COALESCE(NULLIF(:city, ''), city),
		    id_number = COALESCE(NULLIF(:id_number, ''), id_number),
		    customer_type = COALESCE(NULLIF(:customer_type, ''), customer_type),
		    company_name = COALESCE(NULLIF(:company_name, ''), company_name),
		    tax_number = COALESCE(NULLIF(:tax_number, ''), tax_number),
		    notes = COALESCE(NULLIF(:notes, ''), notes),
		    is_active = COALESCE(:is_active, is_active),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = :id AND deleted_at IS NULL
		RETURNING updated_at
	`

	customer.ID = id
	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	err = stmt.Get(&customer.UpdatedAt, customer)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("customer not found")
		}
		return fmt.Errorf("failed to update customer: %w", err)
	}

	return nil
}

func (r *customerRepositoryImpl) Delete(id int) error {
	query := `
		UPDATE customers 
		SET deleted_at = CURRENT_TIMESTAMP 
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("customer not found")
	}

	return nil
}

func (r *customerRepositoryImpl) Count() (int, error) {
	query := `SELECT COUNT(*) FROM customers WHERE deleted_at IS NULL`

	var count int
	err := r.db.Get(&count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count customers: %w", err)
	}

	return count, nil
}

func (r *customerRepositoryImpl) GenerateCustomerCode() (string, error) {
	query := `SELECT generate_customer_code()`

	var code string
	err := r.db.Get(&code, query)
	if err != nil {
		return "", fmt.Errorf("failed to generate customer code: %w", err)
	}

	return code, nil
}
