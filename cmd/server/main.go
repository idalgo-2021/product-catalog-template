// cmd/server/main.go
package main

// @title           Test Project One API
// @version         1.0
// @description     This is a sample server for a product showcase project.
// @termsOfService  http://swagger.io/terms/

// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"go.uber.org/zap"

	"github.com/idalgo-2021/product-catalog-backend/internal/config"
	"github.com/idalgo-2021/product-catalog-backend/internal/handler/auth"
	"github.com/idalgo-2021/product-catalog-backend/internal/handler/order"

	"github.com/idalgo-2021/product-catalog-backend/internal/handler/product"

	"github.com/idalgo-2021/product-catalog-backend/internal/repository/postgres"
	"github.com/idalgo-2021/product-catalog-backend/internal/router"
	"github.com/idalgo-2021/product-catalog-backend/internal/service"
	pkgdb "github.com/idalgo-2021/product-catalog-backend/pkg/db"
	pkglogger "github.com/idalgo-2021/product-catalog-backend/pkg/logger"
)

func main() {

	cfg, err := config.LoadConfig(".env")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := pkglogger.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	dbpool, err := pkgdb.NewPostgresPool(context.Background(), cfg.DatabaseURL, pkgdb.PoolConfig{
		MaxConns:        cfg.DBMaxConns,
		MinConns:        cfg.DBMinConns,
		MaxConnIdleTime: cfg.DBMaxConnIdleTime,
		MaxConnLifetime: cfg.DBMaxConnLifetime,
	})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer dbpool.Close()

	logger.Info("Connected to database")

	userRepo := postgres.NewUserRepo(dbpool, logger)
	productRepo := postgres.NewProductRepo(dbpool, logger)
	orderRepo := postgres.NewOrderRepo(dbpool, logger)

	authService := service.NewAuthService(
		userRepo,
		cfg.JWTSecret,
		cfg.AccessTokenExpiry,
		cfg.RefreshTokenExpiry,
		logger,
	)
	productService := service.NewProductService(productRepo, logger)
	orderService := service.NewOrderService(orderRepo, productRepo, logger)

	authHandler := auth.NewAuthHandler(authService, logger)
	productHandler := product.NewProductHandler(productService, cfg.DefaultListLimit, cfg.MaxListLimit, logger)
	orderHandler := order.NewOrderHandler(orderService, cfg.DefaultListLimit, cfg.MaxListLimit, logger)

	router := router.New(
		authHandler,
		productHandler,
		orderHandler,
		dbpool,
		cfg.JWTSecret,
		logger,
	)

	corsMiddleware := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // In production, specific domains should be listed here (e.g., the frontend domain)
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"}),
		handlers.ExposedHeaders([]string{"Link"}),
		handlers.AllowCredentials(),
	)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.HttpServerPort),
		Handler:      corsMiddleware(router),
		ReadTimeout:  cfg.HttpReadTimeout,
		WriteTimeout: cfg.HttpWriteTimeout,
		IdleTimeout:  cfg.HttpIdleTimeout,
	}

	go func() {
		logger.Info("Starting server", zap.String("port", cfg.HttpServerPort))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}
