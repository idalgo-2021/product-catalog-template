package router

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"

	_ "github.com/idalgo-2021/product-catalog-backend/docs"
	"github.com/idalgo-2021/product-catalog-backend/internal/handler/auth"
	"github.com/idalgo-2021/product-catalog-backend/internal/handler/order"
	"github.com/idalgo-2021/product-catalog-backend/internal/handler/product"
	"github.com/idalgo-2021/product-catalog-backend/internal/middleware"
)

func New(
	authHandler *auth.AuthHandler,
	productHandler *product.ProductHandler,
	orderHandler *order.OrderHandler,
	dbpool Pinger,
	jwtSecret string,
	logger *zap.Logger,
) http.Handler {
	router := mux.NewRouter()

	router.Use(middleware.Recovery(logger))
	router.Use(middleware.Logging(logger))

	// Infrastructure endpoints (no auth required).
	router.HandleFunc("/health", healthHandler).Methods("GET")
	router.HandleFunc("/ready", readyHandler(dbpool)).Methods("GET")

	// Swagger documentation endpoint
	router.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	api := router.PathPrefix("/api/v1").Subrouter()

	// Public endpoints.
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	api.HandleFunc("/auth/refresh", authHandler.Refresh).Methods("POST")
	api.HandleFunc("/auth/validate", authHandler.Validate).Methods("POST")

	api.HandleFunc("/products", productHandler.ListProducts).Methods("GET")
	api.HandleFunc("/products/{product_id}", productHandler.GetProduct).Methods("GET")

	// Protected endpoints — subrouter with empty prefix applies Auth middleware
	// to all routes registered on it, while inheriting the parent path prefix.
	protected := api.PathPrefix("").Subrouter()
	protected.Use(middleware.Auth(jwtSecret, logger))

	protected.HandleFunc("/orders", orderHandler.ListOrders).Methods("GET")
	protected.HandleFunc("/orders", orderHandler.CreateOrder).Methods("POST")
	protected.HandleFunc("/orders/{order_id}", orderHandler.GetOrder).Methods("GET")
	protected.HandleFunc("/orders/{order_id}", orderHandler.UpdateOrder).Methods("PATCH")

	return router
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

type Pinger interface {
	Ping(ctx context.Context) error
}

func readyHandler(db Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if err := db.Ping(r.Context()); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "error",
				"detail": "database unavailable",
			})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
	}
}
