package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/baditaflorin/commonuseragent/internal/api"
	"github.com/baditaflorin/commonuseragent/internal/config"
	"github.com/baditaflorin/commonuseragent/internal/database"
	"github.com/baditaflorin/commonuseragent/internal/web"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.New(
		cfg.Database.Path,
		cfg.Database.MaxOpenConns,
		cfg.Database.MaxIdleConns,
		cfg.Database.ConnMaxLifetime,
	)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Verify database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Printf("Database initialized successfully at %s", cfg.Database.Path)

	// Initialize API handler
	apiHandler := api.NewHandler(db)

	// Initialize web server
	webServer, err := web.NewServer()
	if err != nil {
		log.Fatalf("Failed to initialize web server: %v", err)
	}

	// Setup HTTP router
	mux := http.NewServeMux()

	// Web UI routes
	mux.HandleFunc("/", webServer.Index)

	// API routes
	mux.HandleFunc("/api/desktop", apiHandler.GetRandomDesktop)
	mux.HandleFunc("/api/mobile", apiHandler.GetRandomMobile)
	mux.HandleFunc("/api/random", apiHandler.GetRandom)
	mux.HandleFunc("/api/all/desktop", apiHandler.GetAllDesktop)
	mux.HandleFunc("/api/all/mobile", apiHandler.GetAllMobile)
	mux.HandleFunc("/api/logs", apiHandler.GetRecentRequests)
	mux.HandleFunc("/api/stats", apiHandler.GetStats)
	mux.HandleFunc("/api/health", apiHandler.Health)

	// Apply rate limiting middleware
	handler := api.RateLimitMiddleware(cfg.App.MaxRequests, time.Minute)(mux)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting server on %s (environment: %s)", addr, cfg.App.Environment)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
