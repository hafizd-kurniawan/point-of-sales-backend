package repositories

import (
	"vehicle-showroom-backend/internal/domain/entities"
)

type TransactionRepository interface {
	// Sales Transaction CRUD
	CreateSalesTransaction(transaction *entities.SalesTransaction) error
	GetSalesTransactionByID(id int) (*entities.SalesTransaction, error)
	GetSalesTransactionByNumber(number string) (*entities.SalesTransaction, error)
	GetAllSalesTransactions(limit, offset int, filter *entities.SalesTransactionFilter) ([]entities.SalesTransaction, error)
	UpdateSalesTransaction(id int, transaction *entities.SalesTransaction) error
	DeleteSalesTransaction(id int) error
	CountSalesTransactions(filter *entities.SalesTransactionFilter) (int, error)

	// Purchase Transaction CRUD
	CreatePurchaseTransaction(transaction *entities.PurchaseTransaction) error
	GetPurchaseTransactionByID(id int) (*entities.PurchaseTransaction, error)
	GetPurchaseTransactionByNumber(number string) (*entities.PurchaseTransaction, error)
	GetAllPurchaseTransactions(limit, offset int, filter *entities.PurchaseTransactionFilter) ([]entities.PurchaseTransaction, error)
	UpdatePurchaseTransaction(id int, transaction *entities.PurchaseTransaction) error
	DeletePurchaseTransaction(id int) error
	CountPurchaseTransactions(filter *entities.PurchaseTransactionFilter) (int, error)

	// Statistics & Analytics
	GetTransactionStatistics(salesFilter *entities.SalesTransactionFilter, purchaseFilter *entities.PurchaseTransactionFilter) (*entities.TransactionStatistics, error)
	GetDailyRevenue(year int, month int) ([]entities.DailyRevenue, error)
	GetTopCashiers(limit int, year *int) ([]entities.CashierPerformance, error)
}
