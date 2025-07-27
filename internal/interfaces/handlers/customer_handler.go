package handlers

import (
	"strconv"
	"vehicle-showroom-backend/internal/domain/entities"
	"vehicle-showroom-backend/internal/usecases"
	"vehicle-showroom-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	customerUsecase usecases.CustomerUsecase
}

func NewCustomerHandler(customerUsecase usecases.CustomerUsecase) *CustomerHandler {
	return &CustomerHandler{
		customerUsecase: customerUsecase,
	}
}

func (h *CustomerHandler) CreateCustomer(c *gin.Context) {
	var request entities.CreateCustomerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	customer, err := h.customerUsecase.CreateCustomer(&request)
	if err != nil {
		response.BadRequest(c, "Failed to create customer", err)
		return
	}

	response.Created(c, "Customer created successfully", customer)
}

func (h *CustomerHandler) GetCustomerByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid customer ID", err)
		return
	}

	customer, err := h.customerUsecase.GetCustomerByID(id)
	if err != nil {
		response.NotFound(c, "Customer not found")
		return
	}

	response.Success(c, "Customer retrieved successfully", customer)
}

func (h *CustomerHandler) GetCustomers(c *gin.Context) {
	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	customers, total, err := h.customerUsecase.GetCustomers(page, limit)
	if err != nil {
		response.InternalServerError(c, "Failed to get customers", err)
		return
	}

	// Calculate pagination meta
	totalPages := (total + limit - 1) / limit
	meta := response.PaginationMeta{
		Page:       page,
		Limit:      limit,
		TotalRows:  total,
		TotalPages: totalPages,
	}

	response.Paginated(c, "Customers retrieved successfully", customers, meta)
}

func (h *CustomerHandler) SearchCustomers(c *gin.Context) {
	var request entities.CustomerSearchRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		response.BadRequest(c, "Invalid search parameters", err)
		return
	}

	// Set defaults
	if request.Page < 1 {
		request.Page = 1
	}
	if request.Limit < 1 || request.Limit > 100 {
		request.Limit = 10
	}

	customers, total, err := h.customerUsecase.SearchCustomers(&request)
	if err != nil {
		response.InternalServerError(c, "Failed to search customers", err)
		return
	}

	// Calculate pagination meta
	totalPages := (total + request.Limit - 1) / request.Limit
	meta := response.PaginationMeta{
		Page:       request.Page,
		Limit:      request.Limit,
		TotalRows:  total,
		TotalPages: totalPages,
	}

	response.Paginated(c, "Customer search completed successfully", customers, meta)
}

func (h *CustomerHandler) UpdateCustomer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid customer ID", err)
		return
	}

	var request entities.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid request format", err)
		return
	}

	customer, err := h.customerUsecase.UpdateCustomer(id, &request)
	if err != nil {
		response.BadRequest(c, "Failed to update customer", err)
		return
	}

	response.Success(c, "Customer updated successfully", customer)
}

func (h *CustomerHandler) DeleteCustomer(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		response.BadRequest(c, "Invalid customer ID", err)
		return
	}

	err = h.customerUsecase.DeleteCustomer(id)
	if err != nil {
		response.BadRequest(c, "Failed to delete customer", err)
		return
	}

	response.Success(c, "Customer deleted successfully", nil)
}
