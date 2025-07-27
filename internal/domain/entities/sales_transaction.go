package entities

import (
	"time"
)

type PaymentMethod string

const (
	PaymentCash     PaymentMethod = "cash"
	PaymentTransfer PaymentMethod = "transfer"
	PaymentCheck    PaymentMethod = "check"
	PaymentCredit   PaymentMethod = "credit"
	PaymentMixed    PaymentMethod = "mixed"
)

type TransactionStatus string

const (
	TransactionCompleted TransactionStatus = "completed"
	TransactionCancelled TransactionStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentCompleted PaymentStatus = "completed"
	PaymentPending   PaymentStatus = "pending"
	PaymentOverdue   PaymentStatus = "overdue"
)

// SalesTransaction - Enhanced with installment support
type SalesTransaction struct {
	ID                int               `json:"id" db:"id"`
	TransactionNumber string            `json:"transaction_number" db:"transaction_number"`
	InvoiceNumber     string            `json:"invoice_number" db:"invoice_number"`
	VehicleID         int               `json:"vehicle_id" db:"vehicle_id"`
	CustomerID        int               `json:"customer_id" db:"customer_id"`
	VehiclePrice      float64           `json:"vehicle_price" db:"vehicle_price"`
	TaxAmount         float64           `json:"tax_amount" db:"tax_amount"`
	DiscountAmount    float64           `json:"discount_amount" db:"discount_amount"`
	TotalAmount       float64           `json:"total_amount" db:"total_amount"`
	PaymentMethod     PaymentMethod     `json:"payment_method" db:"payment_method"`
	PaymentReference  string            `json:"payment_reference" db:"payment_reference"`
	TransactionDate   time.Time         `json:"transaction_date" db:"transaction_date"`
	CashierID         int               `json:"cashier_id" db:"cashier_id"`
	Status            TransactionStatus `json:"status" db:"status"`
	Notes             string            `json:"notes" db:"notes"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`

	// Installment fields
	DownPayment       float64       `json:"down_payment" db:"down_payment"`
	RemainingAmount   float64       `json:"remaining_amount" db:"remaining_amount"`
	PaymentStatus     PaymentStatus `json:"payment_status" db:"payment_status"`
	InstallmentPlan   string        `json:"installment_plan" db:"installment_plan"`
	InstallmentMonths int           `json:"installment_months" db:"installment_months"`
	MonthlyPayment    float64       `json:"monthly_payment" db:"monthly_payment"`
	InterestRate      float64       `json:"interest_rate" db:"interest_rate"`
	BankName          string        `json:"bank_name" db:"bank_name"`
	LoanReference     string        `json:"loan_reference" db:"loan_reference"`
	DownPaymentDate   *time.Time    `json:"down_payment_date" db:"down_payment_date"`

	// Relations (populated dengan joins)
	Customer *Customer `json:"customer,omitempty"`
	Vehicle  *Vehicle  `json:"vehicle,omitempty"`
	Cashier  *User     `json:"cashier,omitempty"`
}

// PurchaseTransaction - Sesuai ERD Original (beli dari customer)
type PurchaseTransaction struct {
	ID                int               `json:"id" db:"id"`
	TransactionNumber string            `json:"transaction_number" db:"transaction_number"`
	InvoiceNumber     string            `json:"invoice_number" db:"invoice_number"`
	VehicleID         int               `json:"vehicle_id" db:"vehicle_id"`
	CustomerID        int               `json:"customer_id" db:"customer_id"`
	VehiclePrice      float64           `json:"vehicle_price" db:"vehicle_price"`
	TaxAmount         float64           `json:"tax_amount" db:"tax_amount"`
	TotalAmount       float64           `json:"total_amount" db:"total_amount"`
	PaymentMethod     PaymentMethod     `json:"payment_method" db:"payment_method"`
	PaymentReference  string            `json:"payment_reference" db:"payment_reference"`
	TransactionDate   time.Time         `json:"transaction_date" db:"transaction_date"`
	CashierID         int               `json:"cashier_id" db:"cashier_id"`
	Status            TransactionStatus `json:"status" db:"status"`
	Notes             string            `json:"notes" db:"notes"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`

	// Relations
	Customer *Customer `json:"customer,omitempty"`
	Vehicle  *Vehicle  `json:"vehicle,omitempty"`
	Cashier  *User     `json:"cashier,omitempty"`
}

// Request DTOs
type CreateSalesTransactionRequest struct {
	VehicleID        int           `json:"vehicle_id" validate:"required"`
	CustomerID       int           `json:"customer_id" validate:"required"`
	VehiclePrice     float64       `json:"vehicle_price" validate:"required,gt=0"`
	TaxAmount        float64       `json:"tax_amount,omitempty"`
	DiscountAmount   float64       `json:"discount_amount,omitempty"`
	PaymentMethod    PaymentMethod `json:"payment_method" validate:"required"`
	PaymentReference string        `json:"payment_reference,omitempty"`
	Notes            string        `json:"notes,omitempty"`

	// Installment fields (optional, required for credit/mixed)
	DownPayment       float64 `json:"down_payment,omitempty"`
	InstallmentMonths int     `json:"installment_months,omitempty"`
	InterestRate      float64 `json:"interest_rate,omitempty"`
	BankName          string  `json:"bank_name,omitempty"`
	LoanReference     string  `json:"loan_reference,omitempty"`
}

type CreatePurchaseTransactionRequest struct {
	VehicleID        int           `json:"vehicle_id" validate:"required"`
	CustomerID       int           `json:"customer_id" validate:"required"`
	VehiclePrice     float64       `json:"vehicle_price" validate:"required,gt=0"`
	TaxAmount        float64       `json:"tax_amount,omitempty"`
	PaymentMethod    PaymentMethod `json:"payment_method" validate:"required"`
	PaymentReference string        `json:"payment_reference,omitempty"`
	Notes            string        `json:"notes,omitempty"`
}

type UpdateTransactionRequest struct {
	Status TransactionStatus `json:"status,omitempty"`
	Notes  string            `json:"notes,omitempty"`
}

// Filter DTOs
type SalesTransactionFilter struct {
	CustomerID    *int               `json:"customer_id,omitempty"`
	VehicleID     *int               `json:"vehicle_id,omitempty"`
	CashierID     *int               `json:"cashier_id,omitempty"`
	Status        *TransactionStatus `json:"status,omitempty"`
	PaymentMethod *PaymentMethod     `json:"payment_method,omitempty"`
	DateFrom      *time.Time         `json:"date_from,omitempty"`
	DateTo        *time.Time         `json:"date_to,omitempty"`
	MinAmount     *float64           `json:"min_amount,omitempty"`
	MaxAmount     *float64           `json:"max_amount,omitempty"`
}

type PurchaseTransactionFilter struct {
	CustomerID    *int               `json:"customer_id,omitempty"`
	VehicleID     *int               `json:"vehicle_id,omitempty"`
	CashierID     *int               `json:"cashier_id,omitempty"`
	Status        *TransactionStatus `json:"status,omitempty"`
	PaymentMethod *PaymentMethod     `json:"payment_method,omitempty"`
	DateFrom      *time.Time         `json:"date_from,omitempty"`
	DateTo        *time.Time         `json:"date_to,omitempty"`
	MinAmount     *float64           `json:"min_amount,omitempty"`
	MaxAmount     *float64           `json:"max_amount,omitempty"`
}

// Statistics DTOs
type TransactionStatistics struct {
	// Sales Statistics
	TotalSalesTransactions  int     `json:"total_sales_transactions"`
	TotalSalesRevenue       float64 `json:"total_sales_revenue"`
	CompletedSales          int     `json:"completed_sales"`
	CancelledSales          int     `json:"cancelled_sales"`
	AverageSalesTransaction float64 `json:"average_sales_transaction"`

	// Purchase Statistics
	TotalPurchaseTransactions  int     `json:"total_purchase_transactions"`
	TotalPurchaseCost          float64 `json:"total_purchase_cost"`
	CompletedPurchases         int     `json:"completed_purchases"`
	CancelledPurchases         int     `json:"cancelled_purchases"`
	AveragePurchaseTransaction float64 `json:"average_purchase_transaction"`

	// Profitability
	TotalProfit  float64 `json:"total_profit"`
	ProfitMargin float64 `json:"profit_margin"`

	// Breakdowns
	SalesPaymentMethodBreakdown    map[string]float64   `json:"sales_payment_method_breakdown"`
	PurchasePaymentMethodBreakdown map[string]float64   `json:"purchase_payment_method_breakdown"`
	DailyRevenue                   []DailyRevenue       `json:"daily_revenue"`
	TopCashiers                    []CashierPerformance `json:"top_cashiers"`
}

type DailyRevenue struct {
	Date          string  `json:"date"`
	SalesRevenue  float64 `json:"sales_revenue"`
	PurchaseCost  float64 `json:"purchase_cost"`
	Profit        float64 `json:"profit"`
	SalesCount    int     `json:"sales_count"`
	PurchaseCount int     `json:"purchase_count"`
}

type CashierPerformance struct {
	CashierID                 int     `json:"cashier_id"`
	CashierName               string  `json:"cashier_name"`
	TotalSalesTransactions    int     `json:"total_sales_transactions"`
	TotalPurchaseTransactions int     `json:"total_purchase_transactions"`
	TotalSalesRevenue         float64 `json:"total_sales_revenue"`
	TotalPurchaseCost         float64 `json:"total_purchase_cost"`
	TotalProfit               float64 `json:"total_profit"`
	AverageSalesTicket        float64 `json:"average_sales_ticket"`
	AveragePurchaseTicket     float64 `json:"average_purchase_ticket"`
}

// ==================== INSTALLMENT SYSTEM ENTITIES ====================

type InstallmentStatus string

const (
	InstallmentPending InstallmentStatus = "pending"
	InstallmentPaid    InstallmentStatus = "paid"
	InstallmentOverdue InstallmentStatus = "overdue"
	InstallmentPartial InstallmentStatus = "partial"
)

type PaymentInstallment struct {
	ID                 int               `json:"id" db:"id"`
	SalesTransactionID int               `json:"sales_transaction_id" db:"sales_transaction_id"`
	InstallmentNumber  int               `json:"installment_number" db:"installment_number"`
	DueDate            time.Time         `json:"due_date" db:"due_date"`
	Amount             float64           `json:"amount" db:"amount"`
	PaidAmount         float64           `json:"paid_amount" db:"paid_amount"`
	PaidDate           *time.Time        `json:"paid_date" db:"paid_date"`
	Status             InstallmentStatus `json:"status" db:"status"`
	LateFee            float64           `json:"late_fee" db:"late_fee"`
	CreatedAt          time.Time         `json:"created_at" db:"created_at"`

	// Relations
	SalesTransaction *SalesTransaction `json:"sales_transaction,omitempty"`
}

type PaymentMethodConfig struct {
	ID                        int     `json:"id" db:"id"`
	MethodName                string  `json:"method_name" db:"method_name"`
	DisplayName               string  `json:"display_name" db:"display_name"`
	RequiresDownPayment       bool    `json:"requires_down_payment" db:"requires_down_payment"`
	MinDownPaymentPercentage  float64 `json:"min_down_payment_percentage" db:"min_down_payment_percentage"`
	MaxInstallmentMonths      int     `json:"max_installment_months" db:"max_installment_months"`
	IsActive                  bool    `json:"is_active" db:"is_active"`
	CreatedAt                 time.Time `json:"created_at" db:"created_at"`
}

// ==================== NEW REQUEST/RESPONSE DTOs ====================

type PayInstallmentRequest struct {
	Amount float64 `json:"amount" validate:"required,gt=0"`
	Notes  string  `json:"notes,omitempty"`
}

type PaymentPreviewRequest struct {
	TotalAmount       float64       `json:"total_amount" validate:"required,gt=0"`
	PaymentMethod     PaymentMethod `json:"payment_method" validate:"required"`
	DownPayment       float64       `json:"down_payment,omitempty"`
	InstallmentMonths int           `json:"installment_months,omitempty"`
	InterestRate      float64       `json:"interest_rate,omitempty"`
}

type PaymentPreviewResponse struct {
	TotalAmount        float64                `json:"total_amount"`
	DownPayment        float64                `json:"down_payment"`
	RemainingAmount    float64                `json:"remaining_amount"`
	InstallmentMonths  int                    `json:"installment_months"`
	MonthlyPayment     float64                `json:"monthly_payment"`
	TotalWithInterest  float64                `json:"total_with_interest"`
	InterestRate       float64                `json:"interest_rate"`
	InterestAmount     float64                `json:"interest_amount"`
	InstallmentPreview []InstallmentPreview   `json:"installment_preview"`
	IsValid            bool                   `json:"is_valid"`
	ValidationErrors   []string               `json:"validation_errors,omitempty"`
}

type InstallmentPreview struct {
	InstallmentNumber int       `json:"installment_number"`
	DueDate           time.Time `json:"due_date"`
	Amount            float64   `json:"amount"`
}

type UpdateInstallmentStatusRequest struct {
	Status InstallmentStatus `json:"status" validate:"required"`
	Notes  string            `json:"notes,omitempty"`
}
