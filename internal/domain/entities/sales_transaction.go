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

// SalesTransaction - Sesuai ERD Original
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

// Payment and Installment Request/Response DTOs
type PaymentPreviewRequest struct {
	TotalAmount       float64 `json:"total_amount" binding:"required,gt=0"`
	PaymentMethod     string  `json:"payment_method" binding:"required"`
	DownPayment       float64 `json:"down_payment" binding:"gte=0"`
	InstallmentMonths int     `json:"installment_months" binding:"gte=0"`
	InterestRate      float64 `json:"interest_rate" binding:"gte=0"`
}

type PayInstallmentRequest struct {
	PaymentAmount    float64 `json:"payment_amount" binding:"required,gt=0"`
	PaymentMethod    string  `json:"payment_method" binding:"required"`
	PaymentReference string  `json:"payment_reference"`
	Notes            string  `json:"notes"`
}

type UpdateInstallmentStatusRequest struct {
	Status   string `json:"status" binding:"required"`
	Notes    string `json:"notes"`
	WaivedBy string `json:"waived_by"`
}

type PaymentMethodResponse struct {
	Method            string   `json:"method"`
	DisplayName       string   `json:"display_name"`
	MinDownPayment    *float64 `json:"min_down_payment,omitempty"`
	RequiresReference bool     `json:"requires_reference"`
	Description       string   `json:"description"`
}

type PaymentPreviewResponse struct {
	TotalAmount       float64 `json:"total_amount"`
	DownPayment       float64 `json:"down_payment"`
	RemainingAmount   float64 `json:"remaining_amount"`
	InstallmentMonths int     `json:"installment_months"`
	MonthlyPayment    float64 `json:"monthly_payment"`
	InterestRate      float64 `json:"interest_rate"`
	TotalInterest     float64 `json:"total_interest"`
	TotalWithInterest float64 `json:"total_with_interest"`
}

type InstallmentStatus string

const (
	InstallmentPending InstallmentStatus = "pending"
	InstallmentPaid    InstallmentStatus = "paid"
	InstallmentOverdue InstallmentStatus = "overdue"
	InstallmentWaived  InstallmentStatus = "waived"
)

type Installment struct {
	ID                int               `json:"id" db:"id"`
	TransactionID     int               `json:"transaction_id" db:"transaction_id"`
	InstallmentNumber int               `json:"installment_number" db:"installment_number"`
	DueDate           time.Time         `json:"due_date" db:"due_date"`
	Amount            float64           `json:"amount" db:"amount"`
	PaidAmount        float64           `json:"paid_amount" db:"paid_amount"`
	Status            InstallmentStatus `json:"status" db:"status"`
	PaidAt            *time.Time        `json:"paid_at" db:"paid_at"`
	PaymentMethod     *PaymentMethod    `json:"payment_method" db:"payment_method"`
	PaymentReference  string            `json:"payment_reference" db:"payment_reference"`
	Notes             string            `json:"notes" db:"notes"`
	WaivedBy          *int              `json:"waived_by" db:"waived_by"`
	CreatedAt         time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at" db:"updated_at"`
}

type InstallmentStats struct {
	TotalInstallments     int     `json:"total_installments"`
	PendingCount          int     `json:"pending_count"`
	OverdueCount          int     `json:"overdue_count"`
	PaidCount             int     `json:"paid_count"`
	TotalPendingAmount    float64 `json:"total_pending_amount"`
	TotalOverdueAmount    float64 `json:"total_overdue_amount"`
	OverduePercentage     float64 `json:"overdue_percentage"`
	CollectionRate        float64 `json:"collection_rate"`
}
