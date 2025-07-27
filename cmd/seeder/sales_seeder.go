package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func seedSalesData(db *sqlx.DB) error {
	fmt.Println("🛒 Seeding sales transaction data...")

	// Sample sales transactions
	salesTransactions := []map[string]interface{}{
		{
			"customer_id":        1,
			"vehicle_id":         1,
			"sales_person_id":    3, // cashier1
			"vehicle_price":      850000000.0,
			"discount_amount":    50000000.0,
			"tax_amount":         80000000.0,
			"total_amount":       880000000.0,
			"payment_method":     "bank_transfer",
			"down_payment":       300000000.0,
			"remaining_payment":  580000000.0,
			"installment_months": 24,
			"monthly_payment":    24166667.0,
			"transaction_status": "completed",
			"payment_status":     "partial",
			"transaction_date":   time.Now().AddDate(0, 0, -7),
			"delivery_date":      time.Now().AddDate(0, 0, -5),
			"notes":              "Customer sangat puas dengan pelayanan",
			"special_terms":      "Gratis service 3 bulan pertama",
		},
		{
			"customer_id":        2,
			"vehicle_id":         2,
			"sales_person_id":    3,
			"vehicle_price":      650000000.0,
			"discount_amount":    25000000.0,
			"tax_amount":         62500000.0,
			"total_amount":       687500000.0,
			"payment_method":     "cash",
			"down_payment":       687500000.0,
			"remaining_payment":  0.0,
			"installment_months": 0,
			"monthly_payment":    0.0,
			"transaction_status": "completed",
			"payment_status":     "paid",
			"transaction_date":   time.Now().AddDate(0, 0, -3),
			"delivery_date":      time.Now().AddDate(0, 0, -1),
			"notes":              "Pembayaran cash full",
		},
		{
			"customer_id":        3,
			"vehicle_id":         3,
			"sales_person_id":    3,
			"vehicle_price":      1200000000.0,
			"discount_amount":    100000000.0,
			"tax_amount":         110000000.0,
			"total_amount":       1210000000.0,
			"payment_method":     "credit",
			"down_payment":       400000000.0,
			"remaining_payment":  810000000.0,
			"installment_months": 36,
			"monthly_payment":    22500000.0,
			"transaction_status": "confirmed",
			"payment_status":     "partial",
			"transaction_date":   time.Now().AddDate(0, 0, -1),
			"notes":              "Menunggu proses kredit bank",
			"special_terms":      "Penyerahan setelah kredit approve",
		},
		{
			"customer_id":        4,
			"vehicle_id":         4,
			"sales_person_id":    3,
			"vehicle_price":      450000000.0,
			"discount_amount":    20000000.0,
			"tax_amount":         43000000.0,
			"total_amount":       473000000.0,
			"payment_method":     "bank_transfer",
			"down_payment":       0.0,
			"remaining_payment":  473000000.0,
			"installment_months": 0,
			"monthly_payment":    0.0,
			"transaction_status": "pending",
			"payment_status":     "unpaid",
			"transaction_date":   time.Now(),
			"notes":              "Customer masih mempertimbangkan",
		},
	}

	for _, tx := range salesTransactions {
		query := `
			INSERT INTO sales_transactions (
				customer_id, vehicle_id, sales_person_id,
				vehicle_price, discount_amount, tax_amount, total_amount,
				payment_method, down_payment, remaining_payment,
				installment_months, monthly_payment,
				transaction_status, payment_status,
				transaction_date, delivery_date, notes, special_terms
			) VALUES (
				:customer_id, :vehicle_id, :sales_person_id,
				:vehicle_price, :discount_amount, :tax_amount, :total_amount,
				:payment_method, :down_payment, :remaining_payment,
				:installment_months, :monthly_payment,
				:transaction_status, :payment_status,
				:transaction_date, :delivery_date, :notes, :special_terms
			) RETURNING id`

		var transactionID int
		stmt, err := db.PrepareNamed(query)
		if err != nil {
			return fmt.Errorf("failed to prepare sales transaction query: %w", err)
		}
		err = stmt.Get(&transactionID, tx)
		if err != nil {
			return fmt.Errorf("failed to insert sales transaction: %w", err)
		}
		stmt.Close()

		// Add payment records untuk yang sudah ada down payment
		if tx["down_payment"].(float64) > 0 {
			paymentRecord := map[string]interface{}{
				"transaction_id":   transactionID,
				"payment_amount":   tx["down_payment"],
				"payment_date":     tx["transaction_date"],
				"payment_method":   tx["payment_method"],
				"reference_number": fmt.Sprintf("REF%d%d", transactionID, rand.Intn(10000)),
				"status":           "confirmed",
				"received_by":      3, // cashier1
				"notes":            "Down payment",
			}

			payQuery := `
				INSERT INTO payment_records (
					transaction_id, payment_amount, payment_date,
					payment_method, reference_number, status, received_by, notes
				) VALUES (
					:transaction_id, :payment_amount, :payment_date,
					:payment_method, :reference_number, :status, :received_by, :notes
				)`

			_, err = db.NamedExec(payQuery, paymentRecord)
			if err != nil {
				return fmt.Errorf("failed to insert payment record: %w", err)
			}
		}
	}

	// Seed beberapa payment records tambahan untuk installments
	additionalPayments := []map[string]interface{}{
		{
			"transaction_id":   1,
			"payment_amount":   24166667.0,
			"payment_date":     time.Now().AddDate(0, -1, 0),
			"payment_method":   "bank_transfer",
			"reference_number": "TRF20241201001",
			"status":           "confirmed",
			"received_by":      3,
			"notes":            "Cicilan bulan 1",
		},
		{
			"transaction_id":   1,
			"payment_amount":   24166667.0,
			"payment_date":     time.Now().AddDate(0, 0, -15),
			"payment_method":   "bank_transfer",
			"reference_number": "TRF20241215001",
			"status":           "confirmed",
			"received_by":      3,
			"notes":            "Cicilan bulan 2",
		},
	}

	for _, payment := range additionalPayments {
		payQuery := `
			INSERT INTO payment_records (
				transaction_id, payment_amount, payment_date,
				payment_method, reference_number, status, received_by, notes
			) VALUES (
				:transaction_id, :payment_amount, :payment_date,
				:payment_method, :reference_number, :status, :received_by, :notes
			)`

		_, err := db.NamedExec(payQuery, payment)
		if err != nil {
			return fmt.Errorf("failed to insert additional payment: %w", err)
		}
	}

	fmt.Println("✅ Sales transaction data seeded successfully!")
	return nil
}

func main() {
	// Database connection
	db, err := sqlx.Connect("postgres", "postgres://admin:admin123@localhost:5432/vehicle_showroom?sslmode=disable")
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Seed sales data
	if err := seedSalesData(db); err != nil {
		log.Fatal("Failed to seed sales data:", err)
	}

	fmt.Println("🎉 Sales seeder completed successfully!")
}
