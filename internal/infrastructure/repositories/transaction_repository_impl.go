package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"

	"github.com/jmoiron/sqlx"
)

type transactionRepositoryImpl struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) repositories.TransactionRepository {
	return &transactionRepositoryImpl{db: db}
}

// Sales Transaction Methods
func (r *transactionRepositoryImpl) CreateSalesTransaction(transaction *entities.SalesTransaction) error {
	query := `
		INSERT INTO sales_transactions (
			vehicle_id, customer_id, vehicle_price, tax_amount, discount_amount, 
			total_amount, payment_method, payment_reference, transaction_date, 
			cashier_id, status, notes
		) VALUES (
			:vehicle_id, :customer_id, :vehicle_price, :tax_amount, :discount_amount,
			:total_amount, :payment_method, :payment_reference, :transaction_date,
			:cashier_id, :status, :notes
		) RETURNING id, transaction_number, invoice_number, created_at
	`

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	err = stmt.Get(transaction, transaction)
	if err != nil {
		return fmt.Errorf("failed to create sales transaction: %w", err)
	}

	return nil
}

func (r *transactionRepositoryImpl) GetSalesTransactionByID(id int) (*entities.SalesTransaction, error) {
	query := `
		SELECT 
			st.*,
			c.id as "customer.id", c.customer_code as "customer.customer_code",
			c.full_name as "customer.full_name", c.email as "customer.email",
			c.phone as "customer.phone", c.customer_type as "customer.customer_type",
			v.id as "vehicle.id", v.vehicle_code as "vehicle.vehicle_code",
			v.brand as "vehicle.brand", v.model as "vehicle.model", 
			v.year as "vehicle.year", v.color as "vehicle.color",
			u.id as "cashier.id", u.username as "cashier.username",
			u.full_name as "cashier.full_name", u.email as "cashier.email"
		FROM sales_transactions st
		LEFT JOIN customers c ON st.customer_id = c.id
		LEFT JOIN vehicles v ON st.vehicle_id = v.id
		LEFT JOIN users u ON st.cashier_id = u.id
		WHERE st.id = $1
	`

	var transaction entities.SalesTransaction
	err := r.db.Get(&transaction, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("sales transaction not found")
		}
		return nil, fmt.Errorf("failed to get sales transaction: %w", err)
	}

	return &transaction, nil
}

func (r *transactionRepositoryImpl) GetSalesTransactionByNumber(number string) (*entities.SalesTransaction, error) {
	query := `
		SELECT 
			st.*,
			c.full_name as "customer.full_name", c.customer_code as "customer.customer_code",
			v.brand as "vehicle.brand", v.model as "vehicle.model", v.vehicle_code as "vehicle.vehicle_code",
			u.full_name as "cashier.full_name"
		FROM sales_transactions st
		LEFT JOIN customers c ON st.customer_id = c.id
		LEFT JOIN vehicles v ON st.vehicle_id = v.id
		LEFT JOIN users u ON st.cashier_id = u.id
		WHERE st.transaction_number = $1 OR st.invoice_number = $1
	`

	var transaction entities.SalesTransaction
	err := r.db.Get(&transaction, query, number)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("sales transaction not found")
		}
		return nil, fmt.Errorf("failed to get sales transaction: %w", err)
	}

	return &transaction, nil
}

func (r *transactionRepositoryImpl) GetAllSalesTransactions(limit, offset int, filter *entities.SalesTransactionFilter) ([]entities.SalesTransaction, error) {
	query := `
		SELECT 
			st.id, st.transaction_number, st.invoice_number, st.vehicle_id, st.customer_id,
			st.vehicle_price, st.tax_amount, st.discount_amount, st.total_amount,
			st.payment_method, st.payment_reference, st.transaction_date,
			st.cashier_id, st.status, st.notes, st.created_at,
			c.full_name as "customer.full_name", c.customer_code as "customer.customer_code",
			v.brand as "vehicle.brand", v.model as "vehicle.model", v.vehicle_code as "vehicle.vehicle_code",
			u.full_name as "cashier.full_name"
		FROM sales_transactions st
		LEFT JOIN customers c ON st.customer_id = c.id
		LEFT JOIN vehicles v ON st.vehicle_id = v.id
		LEFT JOIN users u ON st.cashier_id = u.id
		WHERE 1=1
	`

	var conditions []string
	var args []interface{}
	argCount := 0

	if filter != nil {
		if filter.CustomerID != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("st.customer_id = $%d", argCount))
			args = append(args, *filter.CustomerID)
		}
		if filter.VehicleID != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("st.vehicle_id = $%d", argCount))
			args = append(args, *filter.VehicleID)
		}
		if filter.CashierID != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("st.cashier_id = $%d", argCount))
			args = append(args, *filter.CashierID)
		}
		if filter.Status != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("st.status = $%d", argCount))
			args = append(args, *filter.Status)
		}
		if filter.PaymentMethod != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("st.payment_method = $%d", argCount))
			args = append(args, *filter.PaymentMethod)
		}
		if filter.DateFrom != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("st.transaction_date >= $%d", argCount))
			args = append(args, *filter.DateFrom)
		}
		if filter.DateTo != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("st.transaction_date <= $%d", argCount))
			args = append(args, *filter.DateTo)
		}
		if filter.MinAmount != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("st.total_amount >= $%d", argCount))
			args = append(args, *filter.MinAmount)
		}
		if filter.MaxAmount != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("st.total_amount <= $%d", argCount))
			args = append(args, *filter.MaxAmount)
		}
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY st.created_at DESC"

	if limit > 0 {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	if offset > 0 {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, offset)
	}

	var transactions []entities.SalesTransaction
	err := r.db.Select(&transactions, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales transactions: %w", err)
	}

	return transactions, nil
}

func (r *transactionRepositoryImpl) UpdateSalesTransaction(id int, transaction *entities.SalesTransaction) error {
	query := `
		UPDATE sales_transactions 
		SET status = COALESCE(NULLIF(:status, ''), status),
		    notes = COALESCE(NULLIF(:notes, ''), notes)
		WHERE id = :id
	`

	transaction.ID = id
	_, err := r.db.NamedExec(query, transaction)
	if err != nil {
		return fmt.Errorf("failed to update sales transaction: %w", err)
	}

	return nil
}

func (r *transactionRepositoryImpl) DeleteSalesTransaction(id int) error {
	query := "DELETE FROM sales_transactions WHERE id = $1"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete sales transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("sales transaction not found")
	}

	return nil
}

func (r *transactionRepositoryImpl) CountSalesTransactions(filter *entities.SalesTransactionFilter) (int, error) {
	query := "SELECT COUNT(*) FROM sales_transactions WHERE 1=1"

	var conditions []string
	var args []interface{}
	argCount := 0

	if filter != nil {
		if filter.CustomerID != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("customer_id = $%d", argCount))
			args = append(args, *filter.CustomerID)
		}
		if filter.Status != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("status = $%d", argCount))
			args = append(args, *filter.Status)
		}
		if filter.DateFrom != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("transaction_date >= $%d", argCount))
			args = append(args, *filter.DateFrom)
		}
		if filter.DateTo != nil {
			argCount++
			conditions = append(conditions, fmt.Sprintf("transaction_date <= $%d", argCount))
			args = append(args, *filter.DateTo)
		}
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	var count int
	err := r.db.Get(&count, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to count sales transactions: %w", err)
	}

	return count, nil
}

// Purchase Transaction Methods (similar implementation)
func (r *transactionRepositoryImpl) CreatePurchaseTransaction(transaction *entities.PurchaseTransaction) error {
	query := `
		INSERT INTO purchase_transactions (
			vehicle_id, customer_id, vehicle_price, tax_amount, 
			total_amount, payment_method, payment_reference, transaction_date, 
			cashier_id, status, notes
		) VALUES (
			:vehicle_id, :customer_id, :vehicle_price, :tax_amount,
			:total_amount, :payment_method, :payment_reference, :transaction_date,
			:cashier_id, :status, :notes
		) RETURNING id, transaction_number, invoice_number, created_at
	`

	stmt, err := r.db.PrepareNamed(query)
	if err != nil {
		return fmt.Errorf("failed to prepare query: %w", err)
	}
	defer stmt.Close()

	err = stmt.Get(transaction, transaction)
	if err != nil {
		return fmt.Errorf("failed to create purchase transaction: %w", err)
	}

	return nil
}

func (r *transactionRepositoryImpl) GetPurchaseTransactionByID(id int) (*entities.PurchaseTransaction, error) {
	query := `
		SELECT 
			pt.*,
			c.full_name as "customer.full_name", c.customer_code as "customer.customer_code",
			v.brand as "vehicle.brand", v.model as "vehicle.model", v.vehicle_code as "vehicle.vehicle_code",
			u.full_name as "cashier.full_name"
		FROM purchase_transactions pt
		LEFT JOIN customers c ON pt.customer_id = c.id
		LEFT JOIN vehicles v ON pt.vehicle_id = v.id
		LEFT JOIN users u ON pt.cashier_id = u.id
		WHERE pt.id = $1
	`

	var transaction entities.PurchaseTransaction
	err := r.db.Get(&transaction, query, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("purchase transaction not found")
		}
		return nil, fmt.Errorf("failed to get purchase transaction: %w", err)
	}

	return &transaction, nil
}

func (r *transactionRepositoryImpl) GetPurchaseTransactionByNumber(number string) (*entities.PurchaseTransaction, error) {
	query := `
		SELECT 
			pt.*,
			c.full_name as "customer.full_name", c.customer_code as "customer.customer_code",
			v.brand as "vehicle.brand", v.model as "vehicle.model", v.vehicle_code as "vehicle.vehicle_code",
			u.full_name as "cashier.full_name"
		FROM purchase_transactions pt
		LEFT JOIN customers c ON pt.customer_id = c.id
		LEFT JOIN vehicles v ON pt.vehicle_id = v.id
		LEFT JOIN users u ON pt.cashier_id = u.id
		WHERE pt.transaction_number = $1 OR pt.invoice_number = $1
	`

	var transaction entities.PurchaseTransaction
	err := r.db.Get(&transaction, query, number)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("purchase transaction not found")
		}
		return nil, fmt.Errorf("failed to get purchase transaction: %w", err)
	}

	return &transaction, nil
}

func (r *transactionRepositoryImpl) GetAllPurchaseTransactions(limit, offset int, filter *entities.PurchaseTransactionFilter) ([]entities.PurchaseTransaction, error) {
	query := `
		SELECT 
			pt.*,
			c.full_name as "customer.full_name", c.customer_code as "customer.customer_code",
			v.brand as "vehicle.brand", v.model as "vehicle.model", v.vehicle_code as "vehicle.vehicle_code",
			u.full_name as "cashier.full_name"
		FROM purchase_transactions pt
		LEFT JOIN customers c ON pt.customer_id = c.id
		LEFT JOIN vehicles v ON pt.vehicle_id = v.id
		LEFT JOIN users u ON pt.cashier_id = u.id
		WHERE 1=1
	`

	// Apply filters (similar to sales transactions)
	// ... filter implementation

	query += " ORDER BY pt.created_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}

	var transactions []entities.PurchaseTransaction
	err := r.db.Select(&transactions, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase transactions: %w", err)
	}

	return transactions, nil
}

func (r *transactionRepositoryImpl) UpdatePurchaseTransaction(id int, transaction *entities.PurchaseTransaction) error {
	query := `
		UPDATE purchase_transactions 
		SET status = COALESCE(NULLIF(:status, ''), status),
		    notes = COALESCE(NULLIF(:notes, ''), notes)
		WHERE id = :id
	`

	transaction.ID = id
	_, err := r.db.NamedExec(query, transaction)
	if err != nil {
		return fmt.Errorf("failed to update purchase transaction: %w", err)
	}

	return nil
}

func (r *transactionRepositoryImpl) DeletePurchaseTransaction(id int) error {
	query := "DELETE FROM purchase_transactions WHERE id = $1"

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete purchase transaction: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("purchase transaction not found")
	}

	return nil
}

func (r *transactionRepositoryImpl) CountPurchaseTransactions(filter *entities.PurchaseTransactionFilter) (int, error) {
	query := "SELECT COUNT(*) FROM purchase_transactions WHERE 1=1"

	var count int
	err := r.db.Get(&count, query)
	if err != nil {
		return 0, fmt.Errorf("failed to count purchase transactions: %w", err)
	}

	return count, nil
}

// Statistics Methods
func (r *transactionRepositoryImpl) GetTransactionStatistics(salesFilter *entities.SalesTransactionFilter, purchaseFilter *entities.PurchaseTransactionFilter) (*entities.TransactionStatistics, error) {
	stats := &entities.TransactionStatistics{
		SalesPaymentMethodBreakdown:    make(map[string]float64),
		PurchasePaymentMethodBreakdown: make(map[string]float64),
		DailyRevenue:                   []entities.DailyRevenue{},
		TopCashiers:                    []entities.CashierPerformance{},
	}

	// Sales statistics
	salesQuery := `
		SELECT 
			COUNT(*) as total_sales_transactions,
			COALESCE(SUM(total_amount), 0) as total_sales_revenue,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_sales,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_sales,
			COALESCE(AVG(total_amount), 0) as average_sales_transaction
		FROM sales_transactions 
		WHERE 1=1
	`

	err := r.db.Get(stats, salesQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get sales statistics: %w", err)
	}

	// Purchase statistics
	purchaseQuery := `
		SELECT 
			COUNT(*) as total_purchase_transactions,
			COALESCE(SUM(total_amount), 0) as total_purchase_cost,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_purchases,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_purchases,
			COALESCE(AVG(total_amount), 0) as average_purchase_transaction
		FROM purchase_transactions 
		WHERE 1=1
	`

	var purchaseStats struct {
		TotalPurchaseTransactions  int     `db:"total_purchase_transactions"`
		TotalPurchaseCost          float64 `db:"total_purchase_cost"`
		CompletedPurchases         int     `db:"completed_purchases"`
		CancelledPurchases         int     `db:"cancelled_purchases"`
		AveragePurchaseTransaction float64 `db:"average_purchase_transaction"`
	}

	err = r.db.Get(&purchaseStats, purchaseQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get purchase statistics: %w", err)
	}

	stats.TotalPurchaseTransactions = purchaseStats.TotalPurchaseTransactions
	stats.TotalPurchaseCost = purchaseStats.TotalPurchaseCost
	stats.CompletedPurchases = purchaseStats.CompletedPurchases
	stats.CancelledPurchases = purchaseStats.CancelledPurchases
	stats.AveragePurchaseTransaction = purchaseStats.AveragePurchaseTransaction

	// Calculate profit
	stats.TotalProfit = stats.TotalSalesRevenue - stats.TotalPurchaseCost
	if stats.TotalSalesRevenue > 0 {
		stats.ProfitMargin = (stats.TotalProfit / stats.TotalSalesRevenue) * 100
	}

	return stats, nil
}

func (r *transactionRepositoryImpl) GetDailyRevenue(year int, month int) ([]entities.DailyRevenue, error) {
	query := `
		SELECT 
			DATE(transaction_date) as date,
			COALESCE(SUM(CASE WHEN table_name = 'sales' THEN total_amount ELSE 0 END), 0) as sales_revenue,
			COALESCE(SUM(CASE WHEN table_name = 'purchase' THEN total_amount ELSE 0 END), 0) as purchase_cost,
			COUNT(CASE WHEN table_name = 'sales' THEN 1 END) as sales_count,
			COUNT(CASE WHEN table_name = 'purchase' THEN 1 END) as purchase_count
		FROM (
			SELECT transaction_date, total_amount, 'sales' as table_name FROM sales_transactions
			UNION ALL
			SELECT transaction_date, total_amount, 'purchase' as table_name FROM purchase_transactions
		) combined
		WHERE EXTRACT(YEAR FROM transaction_date) = $1 
		AND EXTRACT(MONTH FROM transaction_date) = $2
		GROUP BY DATE(transaction_date)
		ORDER BY date
	`

	var dailyRevenue []entities.DailyRevenue
	err := r.db.Select(&dailyRevenue, query, year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily revenue: %w", err)
	}

	// Calculate profit for each day
	for i := range dailyRevenue {
		dailyRevenue[i].Profit = dailyRevenue[i].SalesRevenue - dailyRevenue[i].PurchaseCost
	}

	return dailyRevenue, nil
}

func (r *transactionRepositoryImpl) GetTopCashiers(limit int, year *int) ([]entities.CashierPerformance, error) {
	query := `
		SELECT 
			u.id as cashier_id,
			u.full_name as cashier_name,
			COALESCE(sales_stats.total_sales_transactions, 0) as total_sales_transactions,
			COALESCE(purchase_stats.total_purchase_transactions, 0) as total_purchase_transactions,
			COALESCE(sales_stats.total_sales_revenue, 0) as total_sales_revenue,
			COALESCE(purchase_stats.total_purchase_cost, 0) as total_purchase_cost,
			COALESCE(sales_stats.average_sales_ticket, 0) as average_sales_ticket,
			COALESCE(purchase_stats.average_purchase_ticket, 0) as average_purchase_ticket
		FROM users u
		LEFT JOIN (
			SELECT 
				cashier_id,
				COUNT(*) as total_sales_transactions,
				SUM(total_amount) as total_sales_revenue,
				AVG(total_amount) as average_sales_ticket
			FROM sales_transactions 
			WHERE status = 'completed'
			GROUP BY cashier_id
		) sales_stats ON u.id = sales_stats.cashier_id
		LEFT JOIN (
			SELECT 
				cashier_id,
				COUNT(*) as total_purchase_transactions,
				SUM(total_amount) as total_purchase_cost,
				AVG(total_amount) as average_purchase_ticket
			FROM purchase_transactions 
			WHERE status = 'completed'
			GROUP BY cashier_id
		) purchase_stats ON u.id = purchase_stats.cashier_id
		WHERE u.role = 'cashier'
		ORDER BY (COALESCE(sales_stats.total_sales_revenue, 0) - COALESCE(purchase_stats.total_purchase_cost, 0)) DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	var cashiers []entities.CashierPerformance
	err := r.db.Select(&cashiers, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get top cashiers: %w", err)
	}

	// Calculate total profit for each cashier
	for i := range cashiers {
		cashiers[i].TotalProfit = cashiers[i].TotalSalesRevenue - cashiers[i].TotalPurchaseCost
	}

	return cashiers, nil
}

// Installment Management Methods
func (r *transactionRepositoryImpl) GetTransactionInstallments(transactionID int) ([]entities.Installment, error) {
	query := `
		SELECT 
			i.id, i.transaction_id, i.installment_number, i.due_date, 
			i.amount, i.paid_amount, i.status, i.paid_at, i.payment_method, 
			i.payment_reference, i.notes, i.waived_by, i.created_at, i.updated_at,
			st.transaction_number, st.invoice_number, st.total_amount as transaction_total,
			c.id as customer_id, c.full_name as customer_name, c.phone as customer_phone, c.email as customer_email
		FROM installments i
		INNER JOIN sales_transactions st ON i.transaction_id = st.id
		INNER JOIN customers c ON st.customer_id = c.id
		WHERE i.transaction_id = $1
		ORDER BY i.installment_number
	`

	rows, err := r.db.Query(query, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction installments: %w", err)
	}
	defer rows.Close()

	var installments []entities.Installment
	for rows.Next() {
		var inst entities.Installment
		var transaction entities.SalesTransaction
		var customer entities.Customer

		err := rows.Scan(
			&inst.ID, &inst.TransactionID, &inst.InstallmentNumber, &inst.DueDate,
			&inst.Amount, &inst.PaidAmount, &inst.Status, &inst.PaidAt, &inst.PaymentMethod,
			&inst.PaymentReference, &inst.Notes, &inst.WaivedBy, &inst.CreatedAt, &inst.UpdatedAt,
			&transaction.TransactionNumber, &transaction.InvoiceNumber, &transaction.TotalAmount,
			&customer.ID, &customer.FullName, &customer.Phone, &customer.Email,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan installment: %w", err)
		}

		inst.Transaction = &transaction
		inst.Customer = &customer
		installments = append(installments, inst)
	}

	return installments, nil
}

func (r *transactionRepositoryImpl) GetInstallmentByID(id int) (*entities.Installment, error) {
	query := `
		SELECT 
			i.id, i.transaction_id, i.installment_number, i.due_date, 
			i.amount, i.paid_amount, i.status, i.paid_at, i.payment_method, 
			i.payment_reference, i.notes, i.waived_by, i.created_at, i.updated_at,
			st.transaction_number, st.invoice_number, st.total_amount as transaction_total,
			c.id as customer_id, c.full_name as customer_name, c.phone as customer_phone, c.email as customer_email
		FROM installments i
		INNER JOIN sales_transactions st ON i.transaction_id = st.id
		INNER JOIN customers c ON st.customer_id = c.id
		WHERE i.id = $1
	`

	var inst entities.Installment
	var transaction entities.SalesTransaction
	var customer entities.Customer

	err := r.db.QueryRow(query, id).Scan(
		&inst.ID, &inst.TransactionID, &inst.InstallmentNumber, &inst.DueDate,
		&inst.Amount, &inst.PaidAmount, &inst.Status, &inst.PaidAt, &inst.PaymentMethod,
		&inst.PaymentReference, &inst.Notes, &inst.WaivedBy, &inst.CreatedAt, &inst.UpdatedAt,
		&transaction.TransactionNumber, &transaction.InvoiceNumber, &transaction.TotalAmount,
		&customer.ID, &customer.FullName, &customer.Phone, &customer.Email,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("installment not found")
		}
		return nil, fmt.Errorf("failed to get installment: %w", err)
	}

	inst.Transaction = &transaction
	inst.Customer = &customer

	return &inst, nil
}

func (r *transactionRepositoryImpl) GetOverdueInstallments(limit, offset int) ([]entities.Installment, int, error) {
	// First update overdue status
	_, err := r.db.Exec("SELECT update_overdue_installments()")
	if err != nil {
		return nil, 0, fmt.Errorf("failed to update overdue installments: %w", err)
	}

	// Count total overdue installments
	countQuery := `
		SELECT COUNT(*) 
		FROM installments i
		INNER JOIN sales_transactions st ON i.transaction_id = st.id
		WHERE i.status IN ('overdue', 'pending') AND i.due_date < CURRENT_DATE
	`
	var total int
	err = r.db.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count overdue installments: %w", err)
	}

	// Get overdue installments with customer info
	query := `
		SELECT 
			i.id, i.transaction_id, i.installment_number, i.due_date, 
			i.amount, i.paid_amount, i.status, i.paid_at, i.payment_method, 
			i.payment_reference, i.notes, i.waived_by, i.created_at, i.updated_at,
			st.transaction_number, st.invoice_number, st.total_amount as transaction_total,
			c.id as customer_id, c.full_name as customer_name, c.phone as customer_phone, c.email as customer_email
		FROM installments i
		INNER JOIN sales_transactions st ON i.transaction_id = st.id
		INNER JOIN customers c ON st.customer_id = c.id
		WHERE i.status IN ('overdue', 'pending') AND i.due_date < CURRENT_DATE
		ORDER BY i.due_date ASC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get overdue installments: %w", err)
	}
	defer rows.Close()

	var installments []entities.Installment
	for rows.Next() {
		var inst entities.Installment
		var transaction entities.SalesTransaction
		var customer entities.Customer

		err := rows.Scan(
			&inst.ID, &inst.TransactionID, &inst.InstallmentNumber, &inst.DueDate,
			&inst.Amount, &inst.PaidAmount, &inst.Status, &inst.PaidAt, &inst.PaymentMethod,
			&inst.PaymentReference, &inst.Notes, &inst.WaivedBy, &inst.CreatedAt, &inst.UpdatedAt,
			&transaction.TransactionNumber, &transaction.InvoiceNumber, &transaction.TotalAmount,
			&customer.ID, &customer.FullName, &customer.Phone, &customer.Email,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan overdue installment: %w", err)
		}

		inst.Transaction = &transaction
		inst.Customer = &customer
		installments = append(installments, inst)
	}

	return installments, total, nil
}

func (r *transactionRepositoryImpl) PayInstallment(installmentID int, paymentAmount float64, paymentMethod, paymentReference, notes string) error {
	query := `
		UPDATE installments 
		SET paid_amount = paid_amount + $2,
			status = CASE 
				WHEN paid_amount + $2 >= amount THEN 'paid'
				ELSE status 
			END,
			paid_at = CASE 
				WHEN paid_amount + $2 >= amount THEN CURRENT_TIMESTAMP
				ELSE paid_at 
			END,
			payment_method = $3,
			payment_reference = $4,
			notes = COALESCE(NULLIF($5, ''), notes),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND status IN ('pending', 'overdue')
	`

	result, err := r.db.Exec(query, installmentID, paymentAmount, paymentMethod, paymentReference, notes)
	if err != nil {
		return fmt.Errorf("failed to pay installment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("installment not found or already paid")
	}

	return nil
}

func (r *transactionRepositoryImpl) UpdateInstallmentStatus(id int, status string, notes string, waivedBy *int) error {
	query := `
		UPDATE installments 
		SET status = $2,
			notes = COALESCE(NULLIF($3, ''), notes),
			waived_by = $4,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	result, err := r.db.Exec(query, id, status, notes, waivedBy)
	if err != nil {
		return fmt.Errorf("failed to update installment status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("installment not found")
	}

	return nil
}

func (r *transactionRepositoryImpl) GetInstallmentStats() (*entities.InstallmentStats, error) {
	// First update overdue status
	_, err := r.db.Exec("SELECT update_overdue_installments()")
	if err != nil {
		return nil, fmt.Errorf("failed to update overdue installments: %w", err)
	}

	query := `
		SELECT 
			COUNT(*) as total_installments,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_count,
			COUNT(CASE WHEN status = 'overdue' THEN 1 END) as overdue_count,
			COUNT(CASE WHEN status = 'paid' THEN 1 END) as paid_count,
			COALESCE(SUM(CASE WHEN status = 'pending' THEN amount - paid_amount END), 0) as total_pending_amount,
			COALESCE(SUM(CASE WHEN status = 'overdue' THEN amount - paid_amount END), 0) as total_overdue_amount
		FROM installments
	`

	var stats entities.InstallmentStats
	err = r.db.QueryRow(query).Scan(
		&stats.TotalInstallments,
		&stats.PendingCount,
		&stats.OverdueCount,
		&stats.PaidCount,
		&stats.TotalPendingAmount,
		&stats.TotalOverdueAmount,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get installment stats: %w", err)
	}

	// Calculate percentages
	if stats.TotalInstallments > 0 {
		stats.OverduePercentage = float64(stats.OverdueCount) / float64(stats.TotalInstallments) * 100
		stats.CollectionRate = float64(stats.PaidCount) / float64(stats.TotalInstallments) * 100
	}

	return &stats, nil
}
