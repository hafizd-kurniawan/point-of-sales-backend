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
