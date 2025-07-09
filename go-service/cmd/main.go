package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"student-report-service/internal/client"
	"student-report-service/internal/config"
	"student-report-service/internal/handlers"
	"student-report-service/internal/pdf"
	"student-report-service/internal/service"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Setup logger
	logger := setupLogger(cfg.Logging)
	logger.Info("Starting Student Report Service")

	// Initialize components
	nodeClient, err := client.NewNodeJSClient(&cfg.NodeJS, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize Node.js client")
	}
	defer nodeClient.Close()

	pdfGenerator, err := pdf.NewGenerator(&cfg.Report)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize PDF generator")
	}

	pdfService := service.NewPDFReportServiceWithConcreteTypes(nodeClient, pdfGenerator, cfg)
	studentPDFHandler := handlers.NewStudentPDFHandler(pdfService)

	// Setup router
	router := setupRouter(studentPDFHandler, logger)

	// Setup CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Configure as needed
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
	})

	// Create server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      c.Handler(router),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.WithField("port", cfg.Server.Port).Info("Starting HTTP server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Setup graceful shutdown
	setupGracefulShutdown(server, logger)
}

func setupLogger(cfg config.LoggingConfig) *logrus.Logger {
	logger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Set log format
	if cfg.Format == "json" {
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	return logger
}

func setupRouter(handler *handlers.StudentPDFHandler, logger *logrus.Logger) *mux.Router {
	router := mux.NewRouter()

	// Add logging middleware
	router.Use(loggingMiddleware(logger))
	router.Use(recoveryMiddleware(logger))

	// Health check endpoint
	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")

	// API routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Student listing endpoint
	api.HandleFunc("/students", handler.GetStudents).Methods("GET")

	// Report generation
	api.HandleFunc("/reports/student/{id:[0-9]+}", handler.CreateStudentPDF).Methods("POST")

	// Cleanup endpoint
	api.HandleFunc("/reports/cleanup", handler.CleanupReports).Methods("POST")

	return router
}

func loggingMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			logger.WithFields(logrus.Fields{
				"method":      r.Method,
				"url":         r.URL.String(),
				"status":      wrapped.statusCode,
				"duration_ms": duration.Milliseconds(),
				"remote_addr": r.RemoteAddr,
				"user_agent":  r.UserAgent(),
			}).Info("Request processed")
		})
	}
}

func recoveryMiddleware(logger *logrus.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.WithFields(logrus.Fields{
						"error":  err,
						"method": r.Method,
						"url":    r.URL.String(),
					}).Error("Panic recovered")

					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

func setupGracefulShutdown(server *http.Server, logger *logrus.Logger) {
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Create a deadline for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("Server forced to shutdown")
	} else {
		logger.Info("Server exited gracefully")
	}
}

// responseWriterWrapper wraps http.ResponseWriter to capture status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
