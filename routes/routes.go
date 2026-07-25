package routes

import (
	"net/http"
	"os"

	"clean-api/database"
	"clean-api/internal/delivery"
	"clean-api/internal/repository"
	"clean-api/internal/usecase"
	"clean-api/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRoutes() *gin.Engine {
	r := gin.Default()

	// CORS Config
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	// Serve static files in uploads folder
	r.Static("/uploads", "./uploads")

	// Initialize dependencies (Clean Architecture)
	userRepo := repository.NewUserRepository(database.DB)
	donationRepo := repository.NewDonationRepository(database.DB)
	inventoryRepo := repository.NewInventoryRepository(database.DB)
	deliveryRepo := repository.NewDeliveryRepository(database.DB)
	branchRepo := repository.NewBranchRepository(database.DB)
	branchReqRepo := repository.NewBranchRequestRepository(database.DB)
	masterItemRepo := repository.NewMasterItemRepository(database.DB)

	userUsecase := usecase.NewUserUsecase(userRepo)
	donationUsecase := usecase.NewDonationUsecase(donationRepo, masterItemRepo)
	inventoryUsecase := usecase.NewInventoryUsecase(inventoryRepo, donationRepo, masterItemRepo)
	deliveryUsecase := usecase.NewDeliveryUsecase(deliveryRepo, inventoryRepo, branchReqRepo)
	branchUsecase := usecase.NewBranchUsecase(branchRepo)
	branchReqUsecase := usecase.NewBranchRequestUsecase(branchReqRepo, deliveryRepo, inventoryRepo, branchRepo)
	masterItemUsecase := usecase.NewMasterItemUsecase(masterItemRepo)

	authHandler := delivery.NewAuthHandler(userUsecase)
	donationHandler := delivery.NewDonationHandler(donationUsecase)
	inventoryHandler := delivery.NewInventoryHandler(inventoryUsecase)
	deliveryHandler := delivery.NewDeliveryHandler(deliveryUsecase)
	branchHandler := delivery.NewBranchHandler(branchUsecase)
	branchReqHandler := delivery.NewBranchRequestHandler(branchReqUsecase)
	masterItemHandler := delivery.NewMasterItemHandler(masterItemUsecase)

	// Health Check Route
	healthHandler := func(c *gin.Context) {
		dbStatus := "connected"
		if database.DB == nil {
			dbStatus = "disconnected"
		} else {
			sqlDB, err := database.DB.DB()
			if err != nil || sqlDB.Ping() != nil {
				dbStatus = "disconnected"
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": dbStatus,
			"port":     os.Getenv("PORT"),
		})
	}
	r.GET("/health", healthHandler)
	r.GET("/api/health", healthHandler)

	// Public Routes
	r.GET("/api/setup", authHandler.Setup)
	r.GET("/api/v1/setup", authHandler.Setup)
	r.POST("/api/login", authHandler.Login)
	r.POST("/api/v1/auth/login", authHandler.Login)

	// Protected Routes (All authenticated users)
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	{
		api.GET("/me", authHandler.GetMe)
		api.GET("/v1/me", authHandler.GetMe)
		api.GET("/v1/branches", branchHandler.List)
		api.GET("/v1/master-items", masterItemHandler.List)

		// Branch Requests
		api.POST("/v1/branch-requests", branchReqHandler.Create)
		api.GET("/v1/branch-requests", branchReqHandler.List)

		// Donations
		api.POST("/v1/donations", donationHandler.Submit)
		api.GET("/v1/donations", donationHandler.List)
		api.GET("/v1/donations/:id", donationHandler.GetByID)
		api.GET("/v1/donations/:id/print", donationHandler.Print)

		// Inventories
		api.GET("/v1/inventory", inventoryHandler.List)

		// Deliveries
		api.GET("/v1/delivery", deliveryHandler.ListAll)
		api.GET("/v1/couriers", authHandler.ListCouriers)

		// --- Role Specific Routes ---

		// Superadmin only
		admin := api.Group("/v1/admin")
		admin.Use(middleware.RoleMiddleware("superadmin"))
		{
			admin.POST("/setup-role", authHandler.SetupRole)
			admin.GET("/users", authHandler.ListUsers)
			admin.DELETE("/users/:id", authHandler.DeleteUser)
			admin.POST("/branches", branchHandler.Create)
			admin.PUT("/branches/:id", branchHandler.Update)
			admin.DELETE("/branches/:id", branchHandler.Delete)
			admin.DELETE("/master-items/:id", masterItemHandler.Delete)
		}

		// Finance only
		finance := api.Group("/v1/finance")
		finance.Use(middleware.RoleMiddleware("superadmin", "finance"))
		{
			finance.PATCH("/donations/:id/verify-fund", donationHandler.VerifyFund)
		}

		// Logistics only
		logistics := api.Group("/v1/logistics")
		logistics.Use(middleware.RoleMiddleware("superadmin", "logistics"))
		{
			logistics.POST("/inventory", inventoryHandler.CreateDirectly)
			logistics.PATCH("/inventory/:id/verify", inventoryHandler.VerifyPhysical)
			logistics.POST("/delivery", deliveryHandler.Create) // Assign delivery
			logistics.PATCH("/branch-requests/:id/approve", branchReqHandler.Approve)
			logistics.PATCH("/branch-requests/:id/reject", branchReqHandler.Reject)
			logistics.POST("/master-items", masterItemHandler.Create)
			logistics.PUT("/master-items/:id", masterItemHandler.Update)
		}

		// Courier / Delivery driver only
		courier := api.Group("/v1/courier")
		courier.Use(middleware.RoleMiddleware("superadmin", "delivery"))
		{
			courier.GET("/delivery", deliveryHandler.ListForCourier)
			courier.PATCH("/delivery/:id/start", deliveryHandler.StartDelivery)
			courier.POST("/delivery/:id/proof", deliveryHandler.UploadProof)
			courier.PATCH("/delivery/:id/complete-without-proof", deliveryHandler.CompleteWithoutProof)
		}
	}

	return r
}
