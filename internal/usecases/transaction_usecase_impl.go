package usecases

import (
	"fmt"
	"math"
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

	// Validate payment method and installment requirements
	err = u.validatePaymentMethod(req, totalAmount)
	if err != nil {
		return nil, err
	}

	// Calculate installment details
	downPayment, remainingAmount, monthlyPayment, paymentStatus := u.calculateInstallmentDetails(req, totalAmount)

	// Set down payment date for credit/mixed payments
	var downPaymentDate *time.Time
	if req.PaymentMethod == entities.PaymentCredit || req.PaymentMethod == entities.PaymentMixed {
		now := time.Now()
		downPaymentDate = &now
	}

	// Create sales transaction
	transaction := &entities.SalesTransaction{
		VehicleID:         req.VehicleID,
		CustomerID:        req.CustomerID,
		VehiclePrice:      req.VehiclePrice,
		TaxAmount:         req.TaxAmount,
		DiscountAmount:    req.DiscountAmount,
		TotalAmount:       totalAmount,
		PaymentMethod:     req.PaymentMethod,
		PaymentReference:  req.PaymentReference,
		TransactionDate:   time.Now(),
		CashierID:         cashierID,
		Status:            entities.TransactionCompleted,
		Notes:             req.Notes,
		DownPayment:       downPayment,
		RemainingAmount:   remainingAmount,
		PaymentStatus:     paymentStatus,
		InstallmentMonths: req.InstallmentMonths,
		MonthlyPayment:    monthlyPayment,
		InterestRate:      req.InterestRate,
		BankName:          req.BankName,
		LoanReference:     req.LoanReference,
		DownPaymentDate:   downPaymentDate,
	}

	// Save transaction
	err = u.transactionRepo.CreateSalesTransaction(transaction)
	if err != nil {
		return nil, fmt.Errorf("failed to create sales transaction: %w", err)
	}

	// Create installment schedule if needed
	if req.InstallmentMonths > 0 {
		err = u.createInstallmentSchedule(transaction.ID, remainingAmount, req.InstallmentMonths, req.InterestRate)
		if err != nil {
			return nil, fmt.Errorf("failed to create installment schedule: %w", err)
		}
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

// ==================== INSTALLMENT MANAGEMENT METHODS ====================

func (u *transactionUsecase) GetTransactionInstallments(transactionID int) ([]entities.PaymentInstallment, error) {
	return u.transactionRepo.GetInstallmentsByTransactionID(transactionID)
}

func (u *transactionUsecase) PayInstallment(installmentID int, req *entities.PayInstallmentRequest) error {
	return u.transactionRepo.PayInstallment(installmentID, req.Amount, req.Notes)
}

func (u *transactionUsecase) GetOverdueInstallments(page, limit int) ([]entities.PaymentInstallment, int, error) {
	offset := (page - 1) * limit
	installments, err := u.transactionRepo.GetOverdueInstallments(limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := u.transactionRepo.CountOverdueInstallments()
	if err != nil {
		return nil, 0, err
	}

	return installments, total, nil
}

func (u *transactionUsecase) UpdateInstallmentStatus(installmentID int, req *entities.UpdateInstallmentStatusRequest) error {
	installment, err := u.transactionRepo.GetInstallmentByID(installmentID)
	if err != nil {
		return err
	}

	installment.Status = req.Status
	return u.transactionRepo.UpdateInstallment(installmentID, installment)
}

// ==================== PAYMENT METHODS & PREVIEW ====================

func (u *transactionUsecase) GetPaymentMethods() ([]entities.PaymentMethodConfig, error) {
	return u.transactionRepo.GetPaymentMethods()
}

func (u *transactionUsecase) GetPaymentPreview(req *entities.PaymentPreviewRequest) (*entities.PaymentPreviewResponse, error) {
	response := &entities.PaymentPreviewResponse{
		TotalAmount:       req.TotalAmount,
		DownPayment:       req.DownPayment,
		InstallmentMonths: req.InstallmentMonths,
		InterestRate:      req.InterestRate,
		IsValid:           true,
		ValidationErrors:  []string{},
	}

	// Validate payment method
	paymentMethod, err := u.transactionRepo.GetPaymentMethodByName(string(req.PaymentMethod))
	if err != nil {
		response.IsValid = false
		response.ValidationErrors = append(response.ValidationErrors, "Invalid payment method")
		return response, nil
	}

	// Validate down payment requirements
	if paymentMethod.RequiresDownPayment {
		minDownPayment := req.TotalAmount * paymentMethod.MinDownPaymentPercentage / 100
		if req.DownPayment < minDownPayment {
			response.IsValid = false
			response.ValidationErrors = append(response.ValidationErrors, 
				fmt.Sprintf("Minimum down payment required: %.2f (%.1f%%)", minDownPayment, paymentMethod.MinDownPaymentPercentage))
		}
	}

	// Validate installment months
	if req.InstallmentMonths > paymentMethod.MaxInstallmentMonths {
		response.IsValid = false
		response.ValidationErrors = append(response.ValidationErrors,
			fmt.Sprintf("Maximum installment months allowed: %d", paymentMethod.MaxInstallmentMonths))
	}

	// Calculate installment details
	if req.InstallmentMonths > 0 {
		response.RemainingAmount = req.TotalAmount - req.DownPayment
		
		// Calculate monthly payment with interest
		monthlyInterestRate := req.InterestRate / 100 / 12
		if monthlyInterestRate > 0 {
			// Use loan payment formula: PMT = P * (r * (1 + r)^n) / ((1 + r)^n - 1)
			numerator := response.RemainingAmount * monthlyInterestRate * math.Pow(1 + monthlyInterestRate, float64(req.InstallmentMonths))
			denominator := math.Pow(1 + monthlyInterestRate, float64(req.InstallmentMonths)) - 1
			response.MonthlyPayment = numerator / denominator
		} else {
			response.MonthlyPayment = response.RemainingAmount / float64(req.InstallmentMonths)
		}

		response.TotalWithInterest = response.DownPayment + (response.MonthlyPayment * float64(req.InstallmentMonths))
		response.InterestAmount = response.TotalWithInterest - req.TotalAmount

		// Generate installment preview
		response.InstallmentPreview = u.generateInstallmentPreview(response.MonthlyPayment, req.InstallmentMonths)
	} else {
		response.RemainingAmount = 0
		response.MonthlyPayment = 0
		response.TotalWithInterest = req.TotalAmount
		response.InterestAmount = 0
	}

	return response, nil
}

// ==================== HELPER METHODS ====================

func (u *transactionUsecase) validatePaymentMethod(req *entities.CreateSalesTransactionRequest, totalAmount float64) error {
	paymentMethod, err := u.transactionRepo.GetPaymentMethodByName(string(req.PaymentMethod))
	if err != nil {
		return fmt.Errorf("invalid payment method: %w", err)
	}

	// Validate down payment requirements
	if paymentMethod.RequiresDownPayment {
		if req.DownPayment <= 0 {
			return fmt.Errorf("down payment is required for %s payment method", paymentMethod.DisplayName)
		}

		minDownPayment := totalAmount * paymentMethod.MinDownPaymentPercentage / 100
		if req.DownPayment < minDownPayment {
			return fmt.Errorf("minimum down payment required: %.2f (%.1f%% of total amount)", 
				minDownPayment, paymentMethod.MinDownPaymentPercentage)
		}

		if req.InstallmentMonths <= 0 {
			return fmt.Errorf("installment months must be specified for %s payment method", paymentMethod.DisplayName)
		}

		if req.InstallmentMonths > paymentMethod.MaxInstallmentMonths {
			return fmt.Errorf("maximum installment months allowed: %d", paymentMethod.MaxInstallmentMonths)
		}
	} else {
		// Cash, transfer, check - no installments allowed
		if req.DownPayment > 0 || req.InstallmentMonths > 0 {
			return fmt.Errorf("%s payment method requires full payment upfront", paymentMethod.DisplayName)
		}
	}

	return nil
}

func (u *transactionUsecase) calculateInstallmentDetails(req *entities.CreateSalesTransactionRequest, totalAmount float64) (downPayment, remainingAmount, monthlyPayment float64, paymentStatus entities.PaymentStatus) {
	if req.PaymentMethod == entities.PaymentCash || req.PaymentMethod == entities.PaymentTransfer || req.PaymentMethod == entities.PaymentCheck {
		// Full payment upfront
		return 0, 0, 0, entities.PaymentCompleted
	}

	// Credit or Mixed payment
	downPayment = req.DownPayment
	remainingAmount = totalAmount - downPayment

	if req.InstallmentMonths > 0 {
		// Calculate monthly payment with interest
		monthlyInterestRate := req.InterestRate / 100 / 12
		if monthlyInterestRate > 0 {
			// Use loan payment formula
			numerator := remainingAmount * monthlyInterestRate * math.Pow(1 + monthlyInterestRate, float64(req.InstallmentMonths))
			denominator := math.Pow(1 + monthlyInterestRate, float64(req.InstallmentMonths)) - 1
			monthlyPayment = numerator / denominator
		} else {
			monthlyPayment = remainingAmount / float64(req.InstallmentMonths)
		}

		return downPayment, remainingAmount, monthlyPayment, entities.PaymentPending
	}

	return downPayment, remainingAmount, 0, entities.PaymentCompleted
}

func (u *transactionUsecase) createInstallmentSchedule(transactionID int, remainingAmount float64, months int, interestRate float64) error {
	monthlyInterestRate := interestRate / 100 / 12
	var monthlyPayment float64

	if monthlyInterestRate > 0 {
		// Calculate monthly payment with interest
		numerator := remainingAmount * monthlyInterestRate * math.Pow(1 + monthlyInterestRate, float64(months))
		denominator := math.Pow(1 + monthlyInterestRate, float64(months)) - 1
		monthlyPayment = numerator / denominator
	} else {
		monthlyPayment = remainingAmount / float64(months)
	}

	var installments []entities.PaymentInstallment
	for i := 1; i <= months; i++ {
		dueDate := time.Now().AddDate(0, i, 0) // Add i months to current date
		
		installment := entities.PaymentInstallment{
			SalesTransactionID: transactionID,
			InstallmentNumber:  i,
			DueDate:            dueDate,
			Amount:             monthlyPayment,
			Status:             entities.InstallmentPending,
			CreatedAt:          time.Now(),
		}
		
		installments = append(installments, installment)
	}

	return u.transactionRepo.CreateInstallments(installments)
}

func (u *transactionUsecase) generateInstallmentPreview(monthlyPayment float64, months int) []entities.InstallmentPreview {
	var preview []entities.InstallmentPreview
	
	for i := 1; i <= months; i++ {
		dueDate := time.Now().AddDate(0, i, 0)
		preview = append(preview, entities.InstallmentPreview{
			InstallmentNumber: i,
			DueDate:           dueDate,
			Amount:            monthlyPayment,
		})
	}
	
	return preview
}
