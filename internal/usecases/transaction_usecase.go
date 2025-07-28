package usecases

import (
	"vehicle-showroom-backend/internal/domain/entities"
)

type TransactionUsecase interface {
	// Sales Transaction Management
	CreateSalesTransaction(req *entities.CreateSalesTransactionRequest, cashierID int) (*entities.SalesTransaction, error)
	GetSalesTransactionByID(id int) (*entities.SalesTransaction, error)
	GetSalesTransactionByNumber(number string) (*entities.SalesTransaction, error)
	GetSalesTransactions(page, limit int, filter *entities.SalesTransactionFilter) ([]entities.SalesTransaction, int, error)
	UpdateSalesTransaction(id int, req *entities.UpdateTransactionRequest) (*entities.SalesTransaction, error)
	CancelSalesTransaction(id int, reason string) error

	// Purchase Transaction Management
	CreatePurchaseTransaction(req *entities.CreatePurchaseTransactionRequest, cashierID int) (*entities.PurchaseTransaction, error)
	GetPurchaseTransactionByID(id int) (*entities.PurchaseTransaction, error)
	GetPurchaseTransactionByNumber(number string) (*entities.PurchaseTransaction, error)
	GetPurchaseTransactions(page, limit int, filter *entities.PurchaseTransactionFilter) ([]entities.PurchaseTransaction, int, error)
	UpdatePurchaseTransaction(id int, req *entities.UpdateTransactionRequest) (*entities.PurchaseTransaction, error)
	CancelPurchaseTransaction(id int, reason string) error

	// Analytics & Reports
	GetTransactionStatistics(salesFilter *entities.SalesTransactionFilter, purchaseFilter *entities.PurchaseTransactionFilter) (*entities.TransactionStatistics, error)
	GetDailyReport(date string) (*DailyReport, error)
	GetMonthlyReport(year int, month int) (*MonthlyReport, error)
	GetCashierPerformance(cashierID int, year *int) (*entities.CashierPerformance, error)

	// Installment Management
	GetTransactionInstallments(transactionID int) ([]entities.Installment, error)
	GetInstallmentByID(id int) (*entities.Installment, error)
	GetOverdueInstallments(page, limit int) ([]entities.Installment, int, error)
	PayInstallment(installmentID int, paymentAmount float64, paymentMethod, paymentReference, notes string) error
	UpdateInstallmentStatus(id int, status string, notes string, waivedBy *int) error
	GetInstallmentStats() (*entities.InstallmentStats, error)
}

// Additional DTOs for reports
type DailyReport struct {
	Date                   string             `json:"date"`
	SalesTransactions      int                `json:"sales_transactions"`
	PurchaseTransactions   int                `json:"purchase_transactions"`
	SalesRevenue           float64            `json:"sales_revenue"`
	PurchaseCost           float64            `json:"purchase_cost"`
	Profit                 float64            `json:"profit"`
	ProfitMargin           float64            `json:"profit_margin"`
	TopCashier             string             `json:"top_cashier"`
	MostSoldBrand          string             `json:"most_sold_brand"`
	PaymentMethodBreakdown map[string]float64 `json:"payment_method_breakdown"`
}

type MonthlyReport struct {
	Year                      int                           `json:"year"`
	Month                     int                           `json:"month"`
	TotalSalesTransactions    int                           `json:"total_sales_transactions"`
	TotalPurchaseTransactions int                           `json:"total_purchase_transactions"`
	TotalSalesRevenue         float64                       `json:"total_sales_revenue"`
	TotalPurchaseCost         float64                       `json:"total_purchase_cost"`
	TotalProfit               float64                       `json:"total_profit"`
	ProfitMargin              float64                       `json:"profit_margin"`
	DailyBreakdown            []entities.DailyRevenue       `json:"daily_breakdown"`
	TopCashiers               []entities.CashierPerformance `json:"top_cashiers"`
	BrandPerformance          []BrandPerformance            `json:"brand_performance"`
}

type BrandPerformance struct {
	Brand         string  `json:"brand"`
	SalesCount    int     `json:"sales_count"`
	PurchaseCount int     `json:"purchase_count"`
	SalesRevenue  float64 `json:"sales_revenue"`
	PurchaseCost  float64 `json:"purchase_cost"`
	Profit        float64 `json:"profit"`
}
