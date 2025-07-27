package main

import (
	"log"
	"vehicle-showroom-backend/internal/config"
	"vehicle-showroom-backend/internal/infrastructure/database"
	infraRepo "vehicle-showroom-backend/internal/infrastructure/repositories"
	"vehicle-showroom-backend/internal/interfaces/handlers"
	"vehicle-showroom-backend/internal/interfaces/middleware"
	"vehicle-showroom-backend/internal/usecases"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Initialize(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize repositories
	userRepo := infraRepo.NewUserRepository(db)
	sessionRepo := infraRepo.NewSessionRepository(db)
	customerRepo := infraRepo.NewCustomerRepository(db)
	vehicleRepo := infraRepo.NewVehicleRepository(db)
	categoryRepo := infraRepo.NewVehicleCategoryRepository(db)
	photoRepo := infraRepo.NewVehiclePhotoRepository(db)
	transactionRepo := infraRepo.NewTransactionRepository(db)

	// Initialize use cases
	authUsecase := usecases.NewAuthUsecase(userRepo, sessionRepo, cfg)
	userUsecase := usecases.NewUserUsecase(userRepo, cfg)
	customerUsecase := usecases.NewCustomerUsecase(customerRepo, cfg)
	vehicleUsecase := usecases.NewVehicleUsecase(vehicleRepo, cfg)
	categoryUsecase := usecases.NewVehicleCategoryUsecase(categoryRepo, cfg)
	photoUsecase := usecases.NewVehiclePhotoUsecase(photoRepo, vehicleRepo, cfg)
	statsUsecase := usecases.NewStatisticsUsecase(vehicleRepo, customerRepo, userRepo, cfg)
	transactionUsecase := usecases.NewTransactionUsecase(
		transactionRepo,
		vehicleRepo,
		customerRepo,
		userRepo,
		cfg,
	)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authUsecase)
	userHandler := handlers.NewUserHandler(userUsecase)
	customerHandler := handlers.NewCustomerHandler(customerUsecase)
	vehicleHandler := handlers.NewVehicleHandler(vehicleUsecase)
	categoryHandler := handlers.NewVehicleCategoryHandler(categoryUsecase)
	photoHandler := handlers.NewVehiclePhotoHandler(photoUsecase)
	statsHandler := handlers.NewStatisticsHandler(statsUsecase)
	transactionHandler := handlers.NewTransactionHandler(transactionUsecase)

	// Initialize router
	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware())

	// Serve static files
	router.Static("/uploads", "./uploads")

	// Routes
	api := router.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/logout", middleware.AuthMiddleware(cfg), authHandler.Logout)
			auth.POST("/refresh", middleware.AuthMiddleware(cfg), authHandler.RefreshToken)
			auth.GET("/me", middleware.AuthMiddleware(cfg), authHandler.GetCurrentUser)
		}

		// Statistics routes
		stats := api.Group("/statistics")
		stats.Use(middleware.AuthMiddleware(cfg))
		{
			stats.GET("/dashboard", statsHandler.GetDashboardStats)
			stats.GET("/vehicles", statsHandler.GetVehicleStats)
			stats.GET("/customers", statsHandler.GetCustomerStats)
		}

		// User routes (admin only)
		users := api.Group("/users")
		users.Use(middleware.AuthMiddleware(cfg))
		users.Use(middleware.RoleMiddleware("admin"))
		{
			users.GET("", userHandler.GetUsers)
			users.POST("", userHandler.CreateUser)
			users.GET("/:id", userHandler.GetUserByID)
			users.PUT("/:id", userHandler.UpdateUser)
			users.DELETE("/:id", userHandler.DeleteUser)
		}

		// Customer routes (admin, cashier)
		customers := api.Group("/customers")
		customers.Use(middleware.AuthMiddleware(cfg))
		customers.Use(middleware.RoleMiddleware("admin", "cashier"))
		{
			customers.GET("", customerHandler.GetCustomers)
			customers.POST("", customerHandler.CreateCustomer)
			customers.GET("/search", customerHandler.SearchCustomers)
			customers.GET("/:id", customerHandler.GetCustomerByID)
			customers.PUT("/:id", customerHandler.UpdateCustomer)
			customers.DELETE("/:id", customerHandler.DeleteCustomer)
		}

		// Vehicle routes - UPDATED WITH FULL CASHIER & MECHANIC ACCESS
		vehicles := api.Group("/vehicles")
		vehicles.Use(middleware.AuthMiddleware(cfg))
		{
			// Read access for all authenticated users
			vehicles.GET("", vehicleHandler.GetVehicles)
			vehicles.GET("/search", vehicleHandler.SearchVehicles)
			vehicles.GET("/:id", vehicleHandler.GetVehicleByID)

			// Create/Update access for admin, cashier, mechanic
			crudVehicles := vehicles.Group("")
			crudVehicles.Use(middleware.RoleMiddleware("admin", "cashier", "mechanic"))
			{
				crudVehicles.POST("", vehicleHandler.CreateVehicle)
				crudVehicles.PUT("/:id", vehicleHandler.UpdateVehicle)

				// NEW: Status and condition updates
				crudVehicles.PATCH("/:id/status", vehicleHandler.UpdateVehicleStatus)
				crudVehicles.PATCH("/:id/condition", vehicleHandler.UpdateVehicleCondition)
			}

			// Delete access (admin only)
			adminVehicles := vehicles.Group("")
			adminVehicles.Use(middleware.RoleMiddleware("admin"))
			{
				adminVehicles.DELETE("/:id", vehicleHandler.DeleteVehicle)
			}
		}

		// Vehicle Photo routes
		photos := api.Group("/photos")
		photos.Use(middleware.AuthMiddleware(cfg))
		{
			photos.GET("/vehicles/:vehicleId", photoHandler.GetVehiclePhotos)
			photos.POST("/vehicles/:vehicleId", photoHandler.UploadPhoto)
			photos.PUT("/vehicles/:vehicleId/:photoId/primary", photoHandler.SetPrimaryPhoto)
			photos.PUT("/:photoId/sort", photoHandler.UpdateSortOrder)
			photos.DELETE("/:photoId", photoHandler.DeletePhoto)
		}

		// Vehicle Category routes
		categories := api.Group("/categories")
		categories.Use(middleware.AuthMiddleware(cfg))
		{
			// All authenticated users can view categories
			categories.GET("", categoryHandler.GetCategories)
			categories.GET("/:id", categoryHandler.GetCategoryByID)

			// Admin only for category management
			adminCategories := categories.Group("")
			adminCategories.Use(middleware.RoleMiddleware("admin"))
			{
				adminCategories.POST("", categoryHandler.CreateCategory)
				adminCategories.PUT("/:id", categoryHandler.UpdateCategory)
				adminCategories.DELETE("/:id", categoryHandler.DeleteCategory)
			}
		}

		// Sales Transaction routes
		sales := api.Group("/sales")
		sales.Use(middleware.AuthMiddleware(cfg))
		sales.Use(middleware.RoleMiddleware("admin", "cashier"))
		{
			sales.POST("/transactions", transactionHandler.CreateSalesTransaction)
			sales.GET("/transactions", transactionHandler.GetSalesTransactions)
			sales.GET("/transactions/:id", transactionHandler.GetSalesTransactionByID)
			sales.GET("/transactions/number/:number", transactionHandler.GetSalesTransactionByNumber)
			sales.PUT("/transactions/:id", transactionHandler.UpdateSalesTransaction)
			sales.POST("/transactions/:id/cancel", transactionHandler.CancelSalesTransaction)

			// NEW: Receipt generation
			sales.POST("/transactions/:id/receipt", transactionHandler.GenerateReceipt)
		}

		// Purchase Transaction routes
		purchase := api.Group("/purchase")
		purchase.Use(middleware.AuthMiddleware(cfg))
		purchase.Use(middleware.RoleMiddleware("admin", "cashier"))
		{
			purchase.POST("/transactions", transactionHandler.CreatePurchaseTransaction)
			purchase.GET("/transactions", transactionHandler.GetPurchaseTransactions)
			purchase.GET("/transactions/:id", transactionHandler.GetPurchaseTransactionByID)
			purchase.GET("/transactions/number/:number", transactionHandler.GetPurchaseTransactionByNumber)
			purchase.PUT("/transactions/:id", transactionHandler.UpdatePurchaseTransaction)
			purchase.POST("/transactions/:id/cancel", transactionHandler.CancelPurchaseTransaction)
		}

		// NEW: Cashier Reports routes
		cashierReports := api.Group("/cashier")
		cashierReports.Use(middleware.AuthMiddleware(cfg))
		cashierReports.Use(middleware.RoleMiddleware("admin", "cashier"))
		{
			cashierReports.GET("/dashboard", transactionHandler.GetCashierDashboard)
			cashierReports.GET("/reports/daily/:date", transactionHandler.GetCashierDailyReport)
			cashierReports.GET("/reports/performance", transactionHandler.GetMyCashierPerformance)
			cashierReports.GET("/reports/summary", transactionHandler.GetCashierSalesSummary)
			cashierReports.GET("/transactions/today", transactionHandler.GetTodayTransactions)
		}

		// NEW: Mechanic routes (using existing vehicle fields)
		mechanic := api.Group("/mechanic")
		mechanic.Use(middleware.AuthMiddleware(cfg))
		mechanic.Use(middleware.RoleMiddleware("admin", "mechanic"))
		{
			mechanic.GET("/dashboard", vehicleHandler.GetMechanicDashboard)
			mechanic.GET("/vehicles/maintenance", vehicleHandler.GetMaintenanceVehicles)
			mechanic.GET("/work-summary", vehicleHandler.GetMechanicWorkSummary)
		}

		// Admin Analytics (full access)
		analytics := api.Group("/analytics")
		analytics.Use(middleware.AuthMiddleware(cfg))
		analytics.Use(middleware.RoleMiddleware("admin"))
		{
			analytics.GET("/statistics", transactionHandler.GetTransactionStatistics)
			analytics.GET("/reports/daily/:date", transactionHandler.GetDailyReport)
			analytics.GET("/reports/monthly/:year/:month", transactionHandler.GetMonthlyReport)
			analytics.GET("/cashier/:id/performance", transactionHandler.GetCashierPerformance)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "OK",
			"message":   "Vehicle Showroom API is running",
			"timestamp": "2025-07-26 00:47:50",
			"version":   "1.0.0",
		})
	})

	log.Printf("Server starting on port %s", cfg.ServerPort)
	log.Fatal(router.Run(":" + cfg.ServerPort))
}
