package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/goesbams/simple-pos-with-terraform/internal/handler"
	"github.com/goesbams/simple-pos-with-terraform/internal/repository"
	"github.com/goesbams/simple-pos-with-terraform/internal/service"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Environment variables
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		dbPort = "5432"
	}
	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		dbUser = "posuser"
	}
	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		dbPassword = "pospassword"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "posdb"
	}

	midtransServerKey := os.Getenv("MIDTRANS_SERVER_KEY")
	midtransClientKey := os.Getenv("MIDTRANS_CLIENT_KEY")
	midtransIsProd := os.Getenv("MIDTRANS_IS_PRODUCTION") == "true"

	// Connect to Database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Printf("Warning: Database connection failed: %v", err)
	} else {
		defer db.Close()
	}

	// Initialize Repositories
	productRepo := repository.NewProductRepository(db)
	cartRepo := repository.NewCartRepository(db)
	txnRepo := repository.NewTransactionRepository(db)

	// Initialize Services
	midtransService := service.NewMidtransService(midtransServerKey, midtransClientKey, midtransIsProd)
	productService := service.NewProductService(productRepo)
	cartService := service.NewCartService(cartRepo, productRepo)
	posService := service.NewPOSService(cartRepo, productRepo, txnRepo, midtransService)

	// Initialize Handlers
	healthHandler := handler.NewHealthHandler(db)
	productHandler := handler.NewProductHandler(productService)
	cartHandler := handler.NewCartHandler(cartService)
	posHandler := handler.NewPOSHandler(posService)

	// Routes
	e.GET("/health", healthHandler.HealthCheck)

	// API v1 Routes
	v1 := e.Group("/api/v1")
	{
		// Products
		v1.GET("/products", productHandler.GetAllProducts)
		v1.GET("/products/:id", productHandler.GetProductByID)
		v1.POST("/products", productHandler.CreateProduct)
		v1.PUT("/products/:id", productHandler.UpdateProduct)
		v1.DELETE("/products/:id", productHandler.SoftDeleteProduct)

		// Cart
		v1.GET("/cart", cartHandler.GetCart)
		v1.POST("/cart/items", cartHandler.AddToCart)
		v1.DELETE("/cart/items/:product_id", cartHandler.RemoveFromCart)
		v1.DELETE("/cart", cartHandler.ClearCart)

		// Checkout & Transactions
		v1.POST("/checkout", posHandler.Checkout)
		v1.GET("/transactions", posHandler.GetAllTransactions)
		v1.GET("/transactions/:id", posHandler.GetTransactionByID)
		v1.POST("/payments/notification", posHandler.MidtransNotificationWebhook)
	}

	log.Printf("Starting Simple POS API server on port %s...", port)
	if err := e.Start(":" + port); err != nil {
		log.Fatalf("Server shutdown: %v", err)
	}
}
