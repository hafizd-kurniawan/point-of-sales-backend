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
