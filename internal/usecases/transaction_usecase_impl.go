package usecases

import (
	"fmt"
	"time"
	"vehicle-showroom-backend/internal/config"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/domain/repositories"
)

type transactionUsecase struct {
	transactionRepo repositories.TransactionRepository
	vehicleRepo     repositories.VehicleRepository
	customerRepo    repositories.CustomerRepository
	userRepo        repositories.UserRepository
	config          *config.Config
}

func NewTransactionUsecase(
	transactionRepo repositories.TransactionRepository,
	vehicleRepo repositories.VehicleRepository,
	customerRepo repositories.CustomerRepository,
	userRepo repositories.UserRepository,
	config *config.Config,
) TransactionUsecase {
	return &transactionUsecase{
		transactionRepo: transactionRepo,
		vehicleRepo:     vehicleRepo,
		customerRepo:    customerRepo,
		userRepo:        userRepo,
		config:          config,
	}
}

// Sales Transaction Methods
func (u *transactionUsecase) CreateSalesTransaction(req *entities.CreateSalesTransactionRequest, cashierID int) (*entities.SalesTransaction, error) {
	// Validate customer exists
	_, err := u.customerRepo.GetByID(req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Validate vehicle exists and available
	vehicle, err := u.vehicleRepo.GetByID(req.VehicleID)
	if err != nil {
		return nil, fmt.Errorf("vehicle not found: %w", err)
	}

	if vehicle.Status != "ready_to_sell" {
		return nil, fmt.Errorf("vehicle is not ready for sale (status: %s)", vehicle.Status)
	}

	// Validate cashier exists and has correct role
	cashier, err := u.userRepo.GetByID(cashierID)
	if err != nil {
		return nil, fmt.Errorf("cashier not found: %w", err)
	}

	if cashier.Role != entities.UserRoleCashier && cashier.Role != entities.UserRoleAdmin {
		return nil, fmt.Errorf("user is not authorized to process sales")
	}

	// Calculate total amount
	totalAmount := req.VehiclePrice + req.TaxAmount - req.DiscountAmount
	if totalAmount <= 0 {
		return nil, fmt.Errorf("total amount must be greater than zero")
	}

	// Create sales transaction
	transaction := &entities.SalesTransaction{
		VehicleID:        req.VehicleID,
		CustomerID:       req.CustomerID,
		VehiclePrice:     req.VehiclePrice,
		TaxAmount:        req.TaxAmount,
		DiscountAmount:   req.DiscountAmount,
		TotalAmount:      totalAmount,
		PaymentMethod:    req.PaymentMethod,
		PaymentReference: req.PaymentReference,
		TransactionDate:  time.Now(),
		CashierID:        cashierID,
		Status:           entities.TransactionCompleted,
		Notes:            req.Notes,
	}

	// Save transaction
	err = u.transactionRepo.CreateSalesTransaction(transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create sales transaction: %w", err)
	}

	// Get full transaction with relations
	return u.transactionRepo.GetSalesTransactionByID(transaction.ID)
}

func (u *transactionUsecase) GetSalesTransactionByID(id int) (*entities.SalesTransaction, error) {
	return u.transactionRepo.GetSalesTransactionByID(id)
}

func (u *transactionUsecase) GetSalesTransactionByNumber(number string) (*entities.SalesTransaction, error) {
	return u.transactionRepo.GetSalesTransactionByNumber(number)
}

func (u *transactionUsecase) GetSalesTransactions(page, limit int, filter *entities.SalesTransactionFilter) ([]entities.SalesTransaction, int, error) {
	offset := (page - 1) * limit

	transactions, err := u.transactionRepo.GetAllSalesTransactions(limit, offset, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get sales transactions: %w", err)
	}

	total, err := u.transactionRepo.CountSalesTransactions(filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count sales transactions: %w", err)
	}

	return transactions, total, nil
}

func (u *transactionUsecase) UpdateSalesTransaction(id int, req *entities.UpdateTransactionRequest) (*entities.SalesTransaction, error) {
	// Get existing transaction
	_, err := u.transactionRepo.GetSalesTransactionByID(id)
	if err != nil {
		return nil, fmt.Errorf("sales transaction not found: %w", err)
	}

	// Prepare update
	update := &entities.SalesTransaction{
		Status: req.Status,
		Notes:  req.Notes,
	}

	// Update transaction
	err = u.transactionRepo.UpdateSalesTransaction(id, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update sales transaction: %w", err)
	}

	// Return updated transaction
	return u.transactionRepo.GetSalesTransactionByID(id)
}

func (u *transactionUsecase) CancelSalesTransaction(id int, reason string) error {
	// Get existing transaction
	existing, err := u.transactionRepo.GetSalesTransactionByID(id)
	if err != nil {
		return fmt.Errorf("sales transaction not found: %w", err)
	}

	if existing.Status == entities.TransactionCancelled {
		return fmt.Errorf("transaction already cancelled")
	}

	// Update to cancelled
	update := &entities.SalesTransaction{
		Status: entities.TransactionCancelled,
		Notes:  fmt.Sprintf("%s | CANCELLED: %s", existing.Notes, reason),
	}

	return u.transactionRepo.UpdateSalesTransaction(id, update)
}

// Purchase Transaction Methods
func (u *transactionUsecase) CreatePurchaseTransaction(req *entities.CreatePurchaseTransactionRequest, cashierID int) (*entities.PurchaseTransaction, error) {
	// Validate customer exists
	_, err := u.customerRepo.GetByID(req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("customer not found: %w", err)
	}

	// Validate vehicle exists
	_, err = u.vehicleRepo.GetByID(req.VehicleID)
	if err != nil {
		return nil, fmt.Errorf("vehicle not found: %w", err)
	}

	// Validate cashier exists and has correct role
	cashier, err := u.userRepo.GetByID(cashierID)
	if err != nil {
		return nil, fmt.Errorf("cashier not found: %w", err)
	}

	if cashier.Role != entities.UserRoleCashier && cashier.Role != entities.UserRoleAdmin {
		return nil, fmt.Errorf("user is not authorized to process purchases")
	}

	// Calculate total amount
	totalAmount := req.VehiclePrice + req.TaxAmount
	if totalAmount <= 0 {
		return nil, fmt.Errorf("total amount must be greater than zero")
	}

	// Create purchase transaction
	transaction := &entities.PurchaseTransaction{
		VehicleID:        req.VehicleID,
		CustomerID:       req.CustomerID,
		VehiclePrice:     req.VehiclePrice,
		TaxAmount:        req.TaxAmount,
		TotalAmount:      totalAmount,
		PaymentMethod:    req.PaymentMethod,
		PaymentReference: req.PaymentReference,
		TransactionDate:  time.Now(),
		CashierID:        cashierID,
		Status:           entities.TransactionCompleted,
		Notes:            req.Notes,
	}

	// Save transaction
	err = u.transactionRepo.CreatePurchaseTransaction(transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create purchase transaction: %w", err)
	}

	// Get full transaction with relations
	return u.transactionRepo.GetPurchaseTransactionByID(transaction.ID)
}

func (u *transactionUsecase) GetPurchaseTransactionByID(id int) (*entities.PurchaseTransaction, error) {
	return u.transactionRepo.GetPurchaseTransactionByID(id)
}

func (u *transactionUsecase) GetPurchaseTransactionByNumber(number string) (*entities.PurchaseTransaction, error) {
	return u.transactionRepo.GetPurchaseTransactionByNumber(number)
}

func (u *transactionUsecase) GetPurchaseTransactions(page, limit int, filter *entities.PurchaseTransactionFilter) ([]entities.PurchaseTransaction, int, error) {
	offset := (page - 1) * limit

	transactions, err := u.transactionRepo.GetAllPurchaseTransactions(limit, offset, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get purchase transactions: %w", err)
	}

	total, err := u.transactionRepo.CountPurchaseTransactions(filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count purchase transactions: %w", err)
	}

	return transactions, total, nil
}

func (u *transactionUsecase) UpdatePurchaseTransaction(id int, req *entities.UpdateTransactionRequest) (*entities.PurchaseTransaction, error) {
	// Get existing transaction
	_, err := u.transactionRepo.GetPurchaseTransactionByID(id)
	if err != nil {
		return nil, fmt.Errorf("purchase transaction not found: %w", err)
	}

	// Prepare update
	update := &entities.PurchaseTransaction{
		Status: req.Status,
		Notes:  req.Notes,
	}

	// Update transaction
	err = u.transactionRepo.UpdatePurchaseTransaction(id, update)
	if err != nil {
		return nil, fmt.Errorf("failed to update purchase transaction: %w", err)
	}

	// Return updated transaction
	return u.transactionRepo.GetPurchaseTransactionByID(id)
}

func (u *transactionUsecase) CancelPurchaseTransaction(id int, reason string) error {
	// Get existing transaction
	existing, err := u.transactionRepo.GetPurchaseTransactionByID(id)
	if err != nil {
		return fmt.Errorf("purchase transaction not found: %w", err)
	}

	if existing.Status == entities.TransactionCancelled {
		return fmt.Errorf("transaction already cancelled")
	}

	// Update to cancelled
	update := &entities.PurchaseTransaction{
		Status: entities.TransactionCancelled,
		Notes:  fmt.Sprintf("%s | CANCELLED: %s", existing.Notes, reason),
	}

	return u.transactionRepo.UpdatePurchaseTransaction(id, update)
}

// Analytics & Reports Methods
func (u *transactionUsecase) GetTransactionStatistics(salesFilter *entities.SalesTransactionFilter, purchaseFilter *entities.PurchaseTransactionFilter) (*entities.TransactionStatistics, error) {
	return u.transactionRepo.GetTransactionStatistics(salesFilter, purchaseFilter)
}

func (u *transactionUsecase) GetDailyReport(date string) (*DailyReport, error) {
	// Parse date
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Create filters for the specific date
	nextDay := parsedDate.Add(24 * time.Hour)
	salesFilter := &entities.SalesTransactionFilter{
		DateFrom: &parsedDate,
		DateTo:   &nextDay,
	}
	purchaseFilter := &entities.PurchaseTransactionFilter{
		DateFrom: &parsedDate,
		DateTo:   &nextDay,
	}

	// Get statistics
	stats, err := u.GetTransactionStatistics(salesFilter, purchaseFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily statistics: %w", err)
	}

	// Build daily report
	report := &DailyReport{
		Date:                   date,
		SalesTransactions:      stats.CompletedSales,
		PurchaseTransactions:   stats.CompletedPurchases,
		SalesRevenue:           stats.TotalSalesRevenue,
		PurchaseCost:           stats.TotalPurchaseCost,
		Profit:                 stats.TotalProfit,
		ProfitMargin:           stats.ProfitMargin,
		PaymentMethodBreakdown: stats.SalesPaymentMethodBreakdown,
	}

	// Get top cashier for the day
	if len(stats.TopCashiers) > 0 {
		report.TopCashier = stats.TopCashiers[0].CashierName
	}

	// Mock most sold brand (would need additional query)
	report.MostSoldBrand = "Toyota"

	return report, nil
}

func (u *transactionUsecase) GetMonthlyReport(year int, month int) (*MonthlyReport, error) {
	// Create date range for the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Nanosecond)

	salesFilter := &entities.SalesTransactionFilter{
		DateFrom: &startDate,
		DateTo:   &endDate,
	}
	purchaseFilter := &entities.PurchaseTransactionFilter{
		DateFrom: &startDate,
		DateTo:   &endDate,
	}

	// Get statistics
	stats, err := u.GetTransactionStatistics(salesFilter, purchaseFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly statistics: %w", err)
	}

	// Get daily breakdown
	dailyRevenue, err := u.transactionRepo.GetDailyRevenue(year, month)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily revenue: %w", err)
	}

	report := &MonthlyReport{
		Year:                      year,
		Month:                     month,
		TotalSalesTransactions:    stats.TotalSalesTransactions,
		TotalPurchaseTransactions: stats.TotalPurchaseTransactions,
		TotalSalesRevenue:         stats.TotalSalesRevenue,
		TotalPurchaseCost:         stats.TotalPurchaseCost,
		TotalProfit:               stats.TotalProfit,
		ProfitMargin:              stats.ProfitMargin,
		DailyBreakdown:            dailyRevenue,
		TopCashiers:               stats.TopCashiers,
		BrandPerformance:          []BrandPerformance{}, // Would need additional query
	}

	return report, nil
}

func (u *transactionUsecase) GetCashierPerformance(cashierID int, year *int) (*entities.CashierPerformance, error) {
	// Create filters
	var salesFilter *entities.SalesTransactionFilter
	var purchaseFilter *entities.PurchaseTransactionFilter

	if year != nil {
		startDate := time.Date(*year, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := startDate.AddDate(1, 0, 0).Add(-time.Nanosecond)

		salesFilter = &entities.SalesTransactionFilter{
			CashierID: &cashierID,
			DateFrom:  &startDate,
			DateTo:    &endDate,
		}
		purchaseFilter = &entities.PurchaseTransactionFilter{
			CashierID: &cashierID,
			DateFrom:  &startDate,
			DateTo:    &endDate,
		}
	} else {
		salesFilter = &entities.SalesTransactionFilter{
			CashierID: &cashierID,
		}
		purchaseFilter = &entities.PurchaseTransactionFilter{
			CashierID: &cashierID,
		}
	}

	// Get statistics
	stats, err := u.GetTransactionStatistics(salesFilter, purchaseFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get cashier performance: %w", err)
	}

	// Get cashier info
	cashier, err := u.userRepo.GetByID(cashierID)
	if err != nil {
		return nil, fmt.Errorf("cashier not found: %w", err)
	}

	performance := &entities.CashierPerformance{
		CashierID:                 cashierID,
		CashierName:               cashier.FullName,
		TotalSalesTransactions:    stats.CompletedSales,
		TotalPurchaseTransactions: stats.CompletedPurchases,
		TotalSalesRevenue:         stats.TotalSalesRevenue,
		TotalPurchaseCost:         stats.TotalPurchaseCost,
		TotalProfit:               stats.TotalProfit,
		AverageSalesTicket:        stats.AverageSalesTransaction,
		AveragePurchaseTicket:     stats.AveragePurchaseTransaction,
	}

	return performance, nil
}
