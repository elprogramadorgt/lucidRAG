package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/config"
	"github.com/elprogramadorgt/lucidRAG/internal/handler"
	"github.com/elprogramadorgt/lucidRAG/internal/middleware"
	"github.com/elprogramadorgt/lucidRAG/internal/rag"
	"github.com/elprogramadorgt/lucidRAG/internal/repository"
	"github.com/elprogramadorgt/lucidRAG/internal/whatsapp"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.New()
	log.Info("Starting lucidRAG service...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Error("Failed to load configuration: %v", err)
		os.Exit(1)
	}

	log.Info("Configuration loaded successfully")
	log.Info("Environment: %s", cfg.Server.Environment)

	// Initialize repositories
	docRepo := repository.NewInMemoryDocumentRepository()
	messageRepo := repository.NewInMemoryMessageRepository()
	sessionRepo := repository.NewInMemorySessionRepository()
	log.Info("Repositories initialized")

	// Initialize services
	whatsappClient := whatsapp.NewClient(&cfg.WhatsApp, log, messageRepo, sessionRepo)
	ragService := rag.NewService(&cfg.RAG, log, docRepo)
	log.Info("Services initialized")

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(log)
	whatsappHandler := handler.NewWhatsAppHandler(whatsappClient, log)
	ragHandler := handler.NewRAGHandler(ragService, log)
	conversationHandler := handler.NewConversationHandler(sessionRepo, messageRepo, log)
	log.Info("Handlers initialized")

	// Setup routes
	mux := http.NewServeMux()

	// Health check
	mux.HandleFunc("/health", healthHandler.HealthCheck)

	// WhatsApp webhook routes
	mux.HandleFunc("/webhook/whatsapp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			whatsappHandler.VerifyWebhook(w, r)
		} else {
			whatsappHandler.HandleWebhook(w, r)
		}
	})

	// RAG API routes
	mux.HandleFunc("/api/v1/rag/query", ragHandler.Query)
	mux.HandleFunc("/api/v1/documents", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if r.URL.Query().Get("id") != "" {
				ragHandler.GetDocument(w, r)
			} else {
				ragHandler.ListDocuments(w, r)
			}
		case http.MethodPost:
			ragHandler.AddDocument(w, r)
		case http.MethodPut:
			ragHandler.UpdateDocument(w, r)
		case http.MethodDelete:
			ragHandler.DeleteDocument(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Conversation API routes
	mux.HandleFunc("/api/v1/conversations", conversationHandler.ListSessions)
	mux.HandleFunc("/api/v1/conversations/session", conversationHandler.GetSession)
	mux.HandleFunc("/api/v1/conversations/messages", conversationHandler.GetMessages)

	// Apply middleware
	handler := middleware.CORS(mux)
	handler = middleware.Logging(log)(handler)
	handler = middleware.Recovery(log)(handler)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Info("Server listening on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("Server forced to shutdown: %v", err)
		os.Exit(1)
	}

	log.Info("Server stopped successfully")
}
