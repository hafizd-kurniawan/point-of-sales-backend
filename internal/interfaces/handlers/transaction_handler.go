package handlers

import (
	"errors"
	"fmt"
	"strconv"
	"time"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/usecases"
	"vehicle-showroom-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	transactionUsecase usecases.TransactionUsecase
}

func NewTransactionHandler(transactionUsecase usecases.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{
		transactionUsecase: transactionUsecase,
	}
}

// ==================== SALES TRANSACTION ENDPOINTS ====================

// @Summary Create new sales transaction
// @Description Create a new vehicle sales transaction
// @Tags Sales
// @Accept json
// @Produce json
// @Param transaction body entities.CreateSalesTransactionRequest true "Sales transaction data"
// @Success 201 {object} response.Response{data=entities.SalesTransaction}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/sales/transactions [post]
func (h *TransactionHandler) CreateSalesTransaction(c *gin.Context) {
	var req entities.CreateSalesTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	// Get cashier ID from context (set by auth middleware)
	cashierID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	cashierIDInt, ok := cashierID.(int)
	if !ok {
		response.InternalServerError(c, "Invalid user ID format", nil)
		return
	}

	transaction, err := h.transactionUsecase.CreateSalesTransaction(&req, cashierIDInt)
	if err != nil {
		response.BadRequest(c, "Failed to create sales transaction", err)
		return
	}

	response.Created(c, "Sales transaction created successfully", transaction)
}

// @Summary Get sales transaction by ID
// @Description Get sales transaction details by ID
// @Tags Sales
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} response.Response{data=entities.SalesTransaction}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/sales/transactions/{id} [get]
func (h *TransactionHandler) GetSalesTransactionByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID", err)
		return
	}

	transaction, err := h.transactionUsecase.GetSalesTransactionByID(id)
	if err != nil {
		response.NotFound(c, "Sales transaction not found")
		return
	}

	response.Success(c, "Sales transaction retrieved successfully", transaction)
}

// @Summary Get sales transaction by number
// @Description Get sales transaction details by transaction number or invoice number
// @Tags Sales
// @Accept json
// @Produce json
// @Param number path string true "Transaction Number or Invoice Number"
// @Success 200 {object} response.Response{data=entities.SalesTransaction}
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/sales/transactions/number/{number} [get]
func (h *TransactionHandler) GetSalesTransactionByNumber(c *gin.Context) {
	number := c.Param("number")
	if number == "" {
		response.BadRequest(c, "Transaction number is required", nil)
		return
	}

	transaction, err := h.transactionUsecase.GetSalesTransactionByNumber(number)
	if err != nil {
		response.NotFound(c, "Sales transaction not found")
		return
	}

	response.Success(c, "Sales transaction retrieved successfully", transaction)
}

// @Summary Get sales transactions list
// @Description Get paginated list of sales transactions with filters
// @Tags Sales
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Param customer_id query int false "Filter by customer ID"
// @Param vehicle_id query int false "Filter by vehicle ID"
// @Param cashier_id query int false "Filter by cashier ID"
// @Param status query string false "Filter by transaction status"
// @Param payment_method query string false "Filter by payment method"
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {object} response.PaginatedResponse{data=[]entities.SalesTransaction}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/sales/transactions [get]
func (h *TransactionHandler) GetSalesTransactions(c *gin.Context) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Parse filters
	filter := &entities.SalesTransactionFilter{}

	if customerID := c.Query("customer_id"); customerID != "" {
		if id, err := strconv.Atoi(customerID); err == nil {
			filter.CustomerID = &id
		}
	}

	if vehicleID := c.Query("vehicle_id"); vehicleID != "" {
		if id, err := strconv.Atoi(vehicleID); err == nil {
			filter.VehicleID = &id
		}
	}

	if cashierID := c.Query("cashier_id"); cashierID != "" {
		if id, err := strconv.Atoi(cashierID); err == nil {
			filter.CashierID = &id
		}
	}

	if status := c.Query("status"); status != "" {
		transactionStatus := entities.TransactionStatus(status)
		filter.Status = &transactionStatus
	}

	if paymentMethod := c.Query("payment_method"); paymentMethod != "" {
		method := entities.PaymentMethod(paymentMethod)
		filter.PaymentMethod = &method
	}

	// Parse dates
	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if parsed, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.DateFrom = &parsed
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if parsed, err := time.Parse("2006-01-02", dateTo); err == nil {
			filter.DateTo = &parsed
		}
	}

	// Parse amount filters
	if minAmount := c.Query("min_amount"); minAmount != "" {
		if amount, err := strconv.ParseFloat(minAmount, 64); err == nil {
			filter.MinAmount = &amount
		}
	}

	if maxAmount := c.Query("max_amount"); maxAmount != "" {
		if amount, err := strconv.ParseFloat(maxAmount, 64); err == nil {
			filter.MaxAmount = &amount
		}
	}

	transactions, total, err := h.transactionUsecase.GetSalesTransactions(page, limit, filter)
	if err != nil {
		response.InternalServerError(c, "Failed to get sales transactions", err)
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := response.PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalRows:  total,
		TotalPages: totalPages,
	}

	response.Paginated(c, "Sales transactions retrieved successfully", transactions, meta)
}

// @Summary Update sales transaction
// @Description Update sales transaction details
// @Tags Sales
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Param transaction body entities.UpdateTransactionRequest true "Update data"
// @Success 200 {object} response.Response{data=entities.SalesTransaction}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/sales/transactions/{id} [put]
func (h *TransactionHandler) UpdateSalesTransaction(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID", err)
		return
	}

	var req entities.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	transaction, err := h.transactionUsecase.UpdateSalesTransaction(id, &req)
	if err != nil {
		response.BadRequest(c, "Failed to update sales transaction", err)
		return
	}

	response.Success(c, "Sales transaction updated successfully", transaction)
}

// @Summary Cancel sales transaction
// @Description Cancel a sales transaction
// @Tags Sales
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Param reason body map[string]string true "Cancellation reason"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/sales/transactions/{id}/cancel [post]
func (h *TransactionHandler) CancelSalesTransaction(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID", err)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Cancellation reason is required", err)
		return
	}

	err = h.transactionUsecase.CancelSalesTransaction(id, req.Reason)
	if err != nil {
		response.BadRequest(c, "Failed to cancel sales transaction", err)
		return
	}

	response.Success(c, "Sales transaction cancelled successfully", nil)
}

// ==================== PURCHASE TRANSACTION ENDPOINTS ====================

// @Summary Create new purchase transaction
// @Description Create a new vehicle purchase transaction
// @Tags Purchase
// @Accept json
// @Produce json
// @Param transaction body entities.CreatePurchaseTransactionRequest true "Purchase transaction data"
// @Success 201 {object} response.Response{data=entities.PurchaseTransaction}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/purchase/transactions [post]
func (h *TransactionHandler) CreatePurchaseTransaction(c *gin.Context) {
	var req entities.CreatePurchaseTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	// Get cashier ID from context
	cashierID, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	cashierIDInt, ok := cashierID.(int)
	if !ok {
		response.InternalServerError(c, "Invalid user ID format", nil)
		return
	}

	transaction, err := h.transactionUsecase.CreatePurchaseTransaction(&req, cashierIDInt)
	if err != nil {
		response.BadRequest(c, "Failed to create purchase transaction", err)
		return
	}

	response.Created(c, "Purchase transaction created successfully", transaction)
}

func (h *TransactionHandler) GetPurchaseTransactionByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID", err)
		return
	}

	transaction, err := h.transactionUsecase.GetPurchaseTransactionByID(id)
	if err != nil {
		msg := fmt.Sprintf("Purchase transaction not found %s", err.Error())
		response.NotFound(c, msg)
		return
	}

	response.Success(c, "Purchase transaction retrieved successfully", transaction)
}

// @Summary Get purchase transactions list
// @Description Get paginated list of purchase transactions with filters
// @Tags Purchase
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.PaginatedResponse{data=[]entities.PurchaseTransaction}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/purchase/transactions [get]
func (h *TransactionHandler) GetPurchaseTransactions(c *gin.Context) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// Parse filters (similar to sales transactions)
	filter := &entities.PurchaseTransactionFilter{}

	transactions, total, err := h.transactionUsecase.GetPurchaseTransactions(page, limit, filter)
	if err != nil {
		response.InternalServerError(c, "Failed to get purchase transactions", err)
		return
	}

	totalPages := (total + limit - 1) / limit
	meta := response.PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalRows:  total,
		TotalPages: totalPages,
	}

	response.Paginated(c, "Purchase transactions retrieved successfully", transactions, meta)
}

// ==================== ANALYTICS & REPORTS ENDPOINTS ====================

// @Summary Get transaction statistics
// @Description Get transaction statistics with optional filters
// @Tags Analytics
// @Accept json
// @Produce json
// @Param date_from query string false "Filter from date (YYYY-MM-DD)"
// @Param date_to query string false "Filter to date (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=entities.TransactionStatistics}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/analytics/statistics [get]
func (h *TransactionHandler) GetTransactionStatistics(c *gin.Context) {
	// Parse filters
	var salesFilter *entities.SalesTransactionFilter
	var purchaseFilter *entities.PurchaseTransactionFilter

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if parsed, err := time.Parse("2006-01-02", dateFrom); err == nil {
			salesFilter = &entities.SalesTransactionFilter{DateFrom: &parsed}
			purchaseFilter = &entities.PurchaseTransactionFilter{DateFrom: &parsed}
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if parsed, err := time.Parse("2006-01-02", dateTo); err == nil {
			if salesFilter == nil {
				salesFilter = &entities.SalesTransactionFilter{}
				purchaseFilter = &entities.PurchaseTransactionFilter{}
			}
			salesFilter.DateTo = &parsed
			purchaseFilter.DateTo = &parsed
		}
	}

	stats, err := h.transactionUsecase.GetTransactionStatistics(salesFilter, purchaseFilter)
	if err != nil {
		response.InternalServerError(c, "Failed to get transaction statistics", err)
		return
	}

	response.Success(c, "Transaction statistics retrieved successfully", stats)
}

// @Summary Get daily report
// @Description Get daily transaction report for a specific date
// @Tags Analytics
// @Accept json
// @Produce json
// @Param date path string true "Date (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=usecases.DailyReport}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/analytics/reports/daily/{date} [get]
func (h *TransactionHandler) GetDailyReport(c *gin.Context) {
	date := c.Param("date")
	if date == "" {
		response.BadRequest(c, "Date is required", nil)
		return
	}

	// Validate date format
	if _, err := time.Parse("2006-01-02", date); err != nil {
		response.BadRequest(c, "Invalid date format, use YYYY-MM-DD", err)
		return
	}

	report, err := h.transactionUsecase.GetDailyReport(date)
	if err != nil {
		response.BadRequest(c, "Failed to get daily report", err)
		return
	}

	response.Success(c, "Daily report retrieved successfully", report)
}

// @Summary Get monthly report
// @Description Get monthly transaction report for a specific year and month
// @Tags Analytics
// @Accept json
// @Produce json
// @Param year path int true "Year"
// @Param month path int true "Month"
// @Success 200 {object} response.Response{data=usecases.MonthlyReport}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/analytics/reports/monthly/{year}/{month} [get]
func (h *TransactionHandler) GetMonthlyReport(c *gin.Context) {
	yearParam := c.Param("year")
	monthParam := c.Param("month")

	year, err := strconv.Atoi(yearParam)
	if err != nil {
		response.BadRequest(c, "Invalid year", err)
		return
	}

	month, err := strconv.Atoi(monthParam)
	if err != nil || month < 1 || month > 12 {
		response.BadRequest(c, "Invalid month (1-12)", err)
		return
	}

	report, err := h.transactionUsecase.GetMonthlyReport(year, month)
	if err != nil {
		response.BadRequest(c, "Failed to get monthly report", err)
		return
	}

	response.Success(c, "Monthly report retrieved successfully", report)
}

// @Summary Get cashier performance
// @Description Get performance statistics for a specific cashier
// @Tags Analytics
// @Accept json
// @Produce json
// @Param id path int true "Cashier ID"
// @Param year query int false "Filter by year"
// @Success 200 {object} response.Response{data=entities.CashierPerformance}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/analytics/cashier/{id}/performance [get]
func (h *TransactionHandler) GetCashierPerformance(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid cashier ID", err)
		return
	}

	var year *int
	if yearParam := c.Query("year"); yearParam != "" {
		if y, err := strconv.Atoi(yearParam); err == nil {
			year = &y
		}
	}

	performance, err := h.transactionUsecase.GetCashierPerformance(id, year)
	if err != nil {
		response.BadRequest(c, "Failed to get cashier performance", err)
		return
	}

	response.Success(c, "Cashier performance retrieved successfully", performance)
}

// Tambahkan method-method ini ke file transaction_handler.go yang sudah ada

// Missing Purchase Transaction Methods
func (h *TransactionHandler) GetPurchaseTransactionByNumber(c *gin.Context) {
	number := c.Param("number")
	if number == "" {
		response.BadRequest(c, "Transaction number is required", errors.New("Transaction number is required"))
		return
	}

	transaction, err := h.transactionUsecase.GetPurchaseTransactionByNumber(number)
	if err != nil {
		response.NotFound(c, "Purchase transaction not found")
		return
	}

	response.Success(c, "Purchase transaction retrieved successfully", transaction)
}

func (h *TransactionHandler) UpdatePurchaseTransaction(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID", err)
		return
	}

	var req entities.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	transaction, err := h.transactionUsecase.UpdatePurchaseTransaction(id, &req)
	if err != nil {
		response.BadRequest(c, "Failed to update purchase transaction", err)
		return
	}

	response.Success(c, "Purchase transaction updated successfully", transaction)
}

func (h *TransactionHandler) CancelPurchaseTransaction(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID", err)
		return
	}

	var req struct {
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Cancellation reason is required", err)
		return
	}

	err = h.transactionUsecase.CancelPurchaseTransaction(id, req.Reason)
	if err != nil {
		response.BadRequest(c, "Failed to cancel purchase transaction", err)
		return
	}

	response.Success(c, "Purchase transaction cancelled successfully", nil)
}

// ==================== PAYMENT AND INSTALLMENT ENDPOINTS ====================

// @Summary Get payment methods
// @Description Get available payment methods with their details
// @Tags Payment
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=[]entities.PaymentMethodResponse}
// @Failure 500 {object} response.Response
// @Router /api/sales/payment-methods [get]
func (h *TransactionHandler) GetPaymentMethods(c *gin.Context) {
	paymentMethods := []entities.PaymentMethodResponse{
		{
			Method:            "cash",
			DisplayName:       "Cash",
			RequiresReference: false,
			Description:       "Cash payment - full amount due immediately",
		},
		{
			Method:            "transfer",
			DisplayName:       "Bank Transfer",
			RequiresReference: true,
			Description:       "Bank transfer payment - requires reference number",
		},
		{
			Method:            "check",
			DisplayName:       "Check",
			RequiresReference: true,
			Description:       "Check payment - requires check number",
		},
		{
			Method:            "credit",
			DisplayName:       "Credit/Installment",
			MinDownPayment:    func() *float64 { v := 0.20; return &v }(), // 20% minimum
			RequiresReference: false,
			Description:       "Credit payment with installments - requires 20% down payment",
		},
		{
			Method:            "mixed",
			DisplayName:       "Mixed Payment",
			RequiresReference: false,
			Description:       "Combination of multiple payment methods",
		},
	}

	response.Success(c, "Payment methods retrieved successfully", paymentMethods)
}

// @Summary Get payment preview
// @Description Calculate payment preview and breakdown
// @Tags Payment
// @Accept json
// @Produce json
// @Param preview body entities.PaymentPreviewRequest true "Payment preview data"
// @Success 200 {object} response.Response{data=entities.PaymentPreviewResponse}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/sales/payment-preview [post]
func (h *TransactionHandler) GetPaymentPreview(c *gin.Context) {
	var req entities.PaymentPreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	// Validate payment method
	switch req.PaymentMethod {
	case "cash", "transfer", "check":
		// For these methods, full payment is required
		req.InstallmentMonths = 0
		req.DownPayment = req.TotalAmount
	case "credit":
		// Credit requires minimum 20% down payment
		minDownPayment := req.TotalAmount * 0.20
		if req.DownPayment < minDownPayment {
			response.BadRequest(c, fmt.Sprintf("Credit payment requires minimum %.2f%% down payment (%.2f)", 20.0, minDownPayment), nil)
			return
		}
		if req.InstallmentMonths <= 0 {
			response.BadRequest(c, "Credit payment requires installment months", nil)
			return
		}
	case "mixed":
		// Mixed allows flexible down payment and installments
		if req.DownPayment < 0 {
			response.BadRequest(c, "Down payment cannot be negative", nil)
			return
		}
	default:
		response.BadRequest(c, "Invalid payment method", nil)
		return
	}

	// Calculate payment details
	remainingAmount := req.TotalAmount - req.DownPayment
	var monthlyPayment, totalInterest, totalWithInterest float64

	if req.InstallmentMonths > 0 && remainingAmount > 0 {
		// Calculate monthly payment with interest
		monthlyInterestRate := req.InterestRate / 100 / 12 // Convert annual percentage to monthly decimal

		if monthlyInterestRate > 0 {
			// Standard loan payment formula: PMT = P * [r(1+r)^n] / [(1+r)^n - 1]
			pow := 1.0
			for i := 0; i < req.InstallmentMonths; i++ {
				pow *= (1 + monthlyInterestRate)
			}
			monthlyPayment = remainingAmount * (monthlyInterestRate * pow) / (pow - 1)
		} else {
			// No interest
			monthlyPayment = remainingAmount / float64(req.InstallmentMonths)
		}

		totalWithInterest = monthlyPayment * float64(req.InstallmentMonths)
		totalInterest = totalWithInterest - remainingAmount
	}

	preview := entities.PaymentPreviewResponse{
		TotalAmount:       req.TotalAmount,
		DownPayment:       req.DownPayment,
		RemainingAmount:   remainingAmount,
		InstallmentMonths: req.InstallmentMonths,
		MonthlyPayment:    monthlyPayment,
		InterestRate:      req.InterestRate,
		TotalInterest:     totalInterest,
		TotalWithInterest: totalWithInterest,
	}

	response.Success(c, "Payment preview calculated successfully", preview)
}

// @Summary Get transaction installments
// @Description Get installment schedule for a transaction
// @Tags Installments
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Success 200 {object} response.Response{data=[]entities.Installment}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/sales/transactions/{id}/installments [get]
func (h *TransactionHandler) GetTransactionInstallments(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID", err)
		return
	}

	// For now, return a mock installment schedule
	// In a real implementation, this would query the database for actual installments
	installments := []entities.Installment{
		{
			ID:                1,
			TransactionID:     id,
			InstallmentNumber: 1,
			DueDate:           time.Now().AddDate(0, 1, 0),
			Amount:            1000000,
			PaidAmount:        0,
			Status:            entities.InstallmentPending,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
	}

	response.Success(c, "Transaction installments retrieved successfully", installments)
}

// @Summary Pay installment
// @Description Process installment payment
// @Tags Installments
// @Accept json
// @Produce json
// @Param id path int true "Transaction ID"
// @Param installmentId path int true "Installment ID"
// @Param payment body entities.PayInstallmentRequest true "Payment data"
// @Success 200 {object} response.Response{data=entities.Installment}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/sales/transactions/{id}/installments/{installmentId}/pay [post]
func (h *TransactionHandler) PayInstallment(c *gin.Context) {
	idParam := c.Param("id")
	installmentIdParam := c.Param("installmentId")

	transactionID, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid transaction ID", err)
		return
	}

	installmentID, err := strconv.Atoi(installmentIdParam)
	if err != nil {
		response.BadRequest(c, "Invalid installment ID", err)
		return
	}

	var req entities.PayInstallmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	// Validate payment method
	switch req.PaymentMethod {
	case "cash", "transfer", "check", "mixed":
		// Valid payment methods for installments
	default:
		response.BadRequest(c, "Invalid payment method for installment", nil)
		return
	}

	// For now, return a mock response
	// In a real implementation, this would:
	// 1. Validate the installment exists and belongs to the transaction
	// 2. Update the installment payment status
	// 3. Record the payment details
	// 4. Update any related transaction status if fully paid

	paymentMethod := entities.PaymentMethod(req.PaymentMethod)
	now := time.Now()

	installment := entities.Installment{
		ID:                installmentID,
		TransactionID:     transactionID,
		InstallmentNumber: 1,
		DueDate:           time.Now().AddDate(0, 1, 0),
		Amount:            req.PaymentAmount,
		PaidAmount:        req.PaymentAmount,
		Status:            entities.InstallmentPaid,
		PaidAt:            &now,
		PaymentMethod:     &paymentMethod,
		PaymentReference:  req.PaymentReference,
		Notes:             req.Notes,
		CreatedAt:         time.Now(),
		UpdatedAt:         now,
	}

	response.Success(c, "Installment payment processed successfully", installment)
}

// @Summary Get overdue installments
// @Description Get list of overdue installments
// @Tags Installments
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} response.PaginatedResponse{data=[]entities.Installment}
// @Failure 400 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/sales/installments/overdue [get]
func (h *TransactionHandler) GetOverdueInstallments(c *gin.Context) {
	// Parse pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	// For now, return mock overdue installments
	// In a real implementation, this would query for installments where:
	// - status = 'pending'
	// - due_date < current_date
	overdueInstallments := []entities.Installment{
		{
			ID:                1,
			TransactionID:     1,
			InstallmentNumber: 1,
			DueDate:           time.Now().AddDate(0, -1, 0), // 1 month overdue
			Amount:            1000000,
			PaidAmount:        0,
			Status:            entities.InstallmentOverdue,
			CreatedAt:         time.Now().AddDate(0, -2, 0),
			UpdatedAt:         time.Now(),
		},
	}

	total := len(overdueInstallments)
	totalPages := (total + limit - 1) / limit
	meta := response.PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalRows:  total,
		TotalPages: totalPages,
	}

	response.Paginated(c, "Overdue installments retrieved successfully", overdueInstallments, meta)
}

// @Summary Update installment status
// @Description Update installment status (waive, etc.)
// @Tags Installments
// @Accept json
// @Produce json
// @Param id path int true "Installment ID"
// @Param status body entities.UpdateInstallmentStatusRequest true "Status update data"
// @Success 200 {object} response.Response{data=entities.Installment}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/sales/installments/{id}/status [patch]
func (h *TransactionHandler) UpdateInstallmentStatus(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid installment ID", err)
		return
	}

	var req entities.UpdateInstallmentStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	// Validate status
	switch req.Status {
	case "pending", "paid", "overdue", "waived":
		// Valid statuses
	default:
		response.BadRequest(c, "Invalid installment status", nil)
		return
	}

	// Get user ID for waived operations
	var waivedBy *int
	if req.Status == "waived" {
		if userID, exists := c.Get("user_id"); exists {
			if uid, ok := userID.(int); ok {
				waivedBy = &uid
			}
		}
	}

	// For now, return a mock response
	// In a real implementation, this would:
	// 1. Validate the installment exists
	// 2. Update the installment status
	// 3. Record who made the change (for waiving)
	// 4. Update timestamps

	now := time.Now()
	installment := entities.Installment{
		ID:                id,
		TransactionID:     1,
		InstallmentNumber: 1,
		DueDate:           time.Now().AddDate(0, 1, 0),
		Amount:            1000000,
		PaidAmount:        0,
		Status:            entities.InstallmentStatus(req.Status),
		Notes:             req.Notes,
		WaivedBy:          waivedBy,
		CreatedAt:         time.Now().AddDate(0, -1, 0),
		UpdatedAt:         now,
	}

	response.Success(c, "Installment status updated successfully", installment)
}

// @Summary Get installment statistics
// @Description Get aggregated statistics for all installments
// @Tags Installments
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=entities.InstallmentStats}
// @Failure 500 {object} response.Response
// @Router /api/sales/installments/stats [get]
func (h *TransactionHandler) GetInstallmentStats(c *gin.Context) {
	// For now, return mock statistics
	// In a real implementation, this would:
	// 1. Query the database for installment statistics
	// 2. Calculate aggregated metrics
	// 3. Return comprehensive statistics

	stats := entities.InstallmentStats{
		TotalInstallments:     150,
		PendingCount:          25,
		OverdueCount:          8,
		PaidCount:             117,
		TotalPendingAmount:    125000000,
		TotalOverdueAmount:    45000000,
		OverduePercentage:     5.33,
		CollectionRate:        78.0,
	}

	response.Success(c, "Installment statistics retrieved successfully", stats)
}
