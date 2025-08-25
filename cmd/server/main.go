package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"
	"github.com/segyhp/billing-engine/internal/config"
	"github.com/segyhp/billing-engine/internal/handler"
	"github.com/segyhp/billing-engine/internal/repository"
	"github.com/segyhp/billing-engine/internal/service"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize Redis
	redisClient := initRedis(cfg)
	defer redisClient.Close()

	//Initialize repositories
	loanRepo := repository.NewLoanRepository(db)
	paymentRepo := repository.NewPaymentRepository(db)

	//Initialize service
	billingService := service.NewBillingService(loanRepo, paymentRepo, redisClient, cfg)
	billingHandler := handler.NewBillingHandler(billingService, cfg)
	healthHandler := handler.NewHealthHandler(db, redisClient)

	// Setup routes
	router := setupRoutes(billingHandler, healthHandler)

	// Start server
	server := &http.Server{
		Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

func initDB(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", cfg.Database.DSN())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	return db, nil
}

func initRedis(cfg *config.Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Host + ":" + cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
}

func setupRoutes(billingHandler *handler.BillingHandler, healthHandler *handler.HealthHandler) *mux.Router {
	router := mux.NewRouter()

	// Health check
	router.HandleFunc("/health", healthHandler.Health).Methods("GET")
	router.HandleFunc("/health/ready", healthHandler.Ready).Methods("GET")

	/// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/loans", billingHandler.CreateLoan).Methods("POST")
	api.HandleFunc("/loans/{loanId}/outstanding", billingHandler.GetOutstanding).Methods("GET")
	api.HandleFunc("/loans/{loanId}/delinquent", billingHandler.IsDelinquent).Methods("GET")
	api.HandleFunc("/loans/{loanId}/payment", billingHandler.MakePayment).Methods("POST")

	return router
}
