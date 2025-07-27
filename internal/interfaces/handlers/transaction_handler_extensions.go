package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GenerateReceipt generates receipt for sales transaction
func (h *TransactionHandler) GenerateReceipt(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "Invalid transaction ID",
			"error":   err.Error(),
		})
		return
	}

	// Get transaction details
	transaction, err := h.transactionUsecase.GetSalesTransactionByID(id)
	if err != nil {
		c.JSON(404, gin.H{
			"success": false,
			"message": "Transaction not found",
			"error":   err.Error(),
		})
		return
	}

	// Generate receipt data
	receipt := map[string]interface{}{
		"transaction_id":     transaction.ID,
		"transaction_number": transaction.TransactionNumber,
		"transaction_date":   transaction.TransactionDate,
		"transaction_type":   "sales",
		"customer":           transaction.Customer,
		"vehicle":            transaction.Vehicle,
		"cashier":            transaction.CashierID,
		"vehicle_price":      transaction.VehiclePrice,
		"tax_amount":         transaction.TaxAmount,
		"discount_amount":    transaction.DiscountAmount,
		"total_amount":       transaction.TotalAmount,
		"payment_method":     transaction.PaymentMethod,
		"payment_reference":  transaction.PaymentReference,
		"notes":              transaction.Notes,
		"generated_at":       time.Now(),
		"receipt_number":     fmt.Sprintf("RCP-%s", transaction.TransactionNumber),
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Receipt generated successfully",
		"data":    receipt,
	})
}

// GetCashierDashboard gets dashboard for cashiers
func (h *TransactionHandler) GetCashierDashboard(c *gin.Context) {
	cashierID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	// Simple dashboard data
	dashboard := map[string]interface{}{
		"cashier_id":            cashierID,
		"date":                  time.Now().Format("2006-01-02"),
		"today_sales_count":     0, // Would need to implement proper counting
		"today_purchase_count":  0,
		"today_sales_amount":    0.0,
		"today_purchase_amount": 0.0,
		"recent_transactions":   []interface{}{},
		"last_updated":          time.Now(),
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Cashier dashboard retrieved successfully",
		"data":    dashboard,
	})
}

// GetCashierDailyReport gets daily report for cashier
func (h *TransactionHandler) GetCashierDailyReport(c *gin.Context) {
	dateParam := c.Param("date")

	cashierID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	report := map[string]interface{}{
		"cashier_id":     cashierID,
		"date":           dateParam,
		"sales_count":    0,
		"purchase_count": 0,
		"sales":          []interface{}{},
		"purchases":      []interface{}{},
		"generated_at":   time.Now(),
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Daily report retrieved successfully",
		"data":    report,
	})
}

// GetMyCashierPerformance gets performance for current cashier
func (h *TransactionHandler) GetMyCashierPerformance(c *gin.Context) {
	cashierID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	var year *int
	if yearParam := c.Query("year"); yearParam != "" {
		if y, err := strconv.Atoi(yearParam); err == nil {
			year = &y
		}
	}

	performance, err := h.transactionUsecase.GetCashierPerformance(cashierID.(int), year)
	if err != nil {
		c.JSON(400, gin.H{
			"success": false,
			"message": "Failed to get performance",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Performance retrieved successfully",
		"data":    performance,
	})
}

// GetCashierSalesSummary gets sales summary for cashier
func (h *TransactionHandler) GetCashierSalesSummary(c *gin.Context) {
	cashierID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	// Default to current month if no dates provided
	if dateFrom == "" || dateTo == "" {
		now := time.Now()
		dateFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		dateTo = now.Format("2006-01-02")
	}

	summary := map[string]interface{}{
		"cashier_id":   cashierID,
		"date_from":    dateFrom,
		"date_to":      dateTo,
		"total_sales":  0,
		"total_amount": 0.0,
		"summary":      "Basic summary data",
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Sales summary retrieved successfully",
		"data":    summary,
	})
}

// GetTodayTransactions gets today's transactions for cashier
func (h *TransactionHandler) GetTodayTransactions(c *gin.Context) {
	cashierID, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	today := time.Now().Format("2006-01-02")

	result := map[string]interface{}{
		"cashier_id":         cashierID,
		"date":               today,
		"sales":              []interface{}{},
		"purchases":          []interface{}{},
		"total_transactions": 0,
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "Today's transactions retrieved successfully",
		"data":    result,
	})
}
