package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	convApp "github.com/elprogramadorgt/lucidRAG/internal/application/conversation"
	docApp "github.com/elprogramadorgt/lucidRAG/internal/application/document"
	userApp "github.com/elprogramadorgt/lucidRAG/internal/application/user"
	"github.com/elprogramadorgt/lucidRAG/internal/application/whatsapp"
	"github.com/elprogramadorgt/lucidRAG/internal/config"
	"github.com/elprogramadorgt/lucidRAG/internal/repository/mongo"
	"github.com/elprogramadorgt/lucidRAG/internal/transport/http/middleware"
	authHandler "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/auth"
	conversationHandler "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/conversation"
	documentHandler "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/document"
	ragHandler "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/rag"
	whatsappHandler "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/whatsapp"
	"github.com/elprogramadorgt/lucidRAG/pkg/chunker"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/elprogramadorgt/lucidRAG/pkg/openai"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logOpts := logger.Options{
		Level: "info",
		JSON:  cfg.Server.Environment == "production",
	}
	if cfg.Server.Environment == "development" {
		logOpts.Level = "debug"
	}

	log := logger.New(logOpts)
	log.Info("Starting lucidRAG service", "environment", cfg.Server.Environment)

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()

	mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
	)

	dbClient, err := mongo.NewClient(ctx, mongoURI, cfg.Database.Name)
	if err != nil {
		log.Error("Failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	log.Info("Connected to MongoDB", "host", cfg.Database.Host, "database", cfg.Database.Name)

	// Initialize repositories
	whatsappRepo := mongo.NewWhatsappRepo(dbClient)
	documentRepo := mongo.NewDocumentRepo(dbClient)
	chunkRepo := mongo.NewChunkRepo(dbClient)
	userRepo := mongo.NewUserRepo(dbClient)
	conversationRepo := mongo.NewConversationRepo(dbClient)
	messageRepo := mongo.NewMessageRepo(dbClient)

	// Initialize OpenAI client (optional - RAG will be disabled if not configured)
	var openaiClient *openai.Client
	if cfg.RAG.OpenAIAPIKey != "" {
		openaiClient = openai.NewClient(cfg.RAG.OpenAIAPIKey)
		log.Info("OpenAI client initialized for RAG")
	} else {
		log.Info("OPENAI_API_KEY not set - RAG functionality will be disabled")
	}

	// Initialize chunker for document processing
	textChunker := chunker.New(cfg.RAG.ChunkSize, cfg.RAG.ChunkOverlap)

	// Initialize services
	whatsappSvc := whatsapp.NewService(whatsappRepo)
	documentSvc := docApp.NewService(docApp.ServiceConfig{
		Repo:           documentRepo,
		ChunkRepo:      chunkRepo,
		OpenAIClient:   openaiClient,
		Chunker:        textChunker,
		EmbeddingModel: cfg.RAG.EmbeddingModel,
		ModelName:      cfg.RAG.ModelName,
	})
	userSvc := userApp.NewService(userApp.ServiceConfig{
		Repo:      userRepo,
		JWTSecret: cfg.Auth.JWTSecret,
		JWTExpiry: time.Duration(cfg.Auth.JWTExpiryHours) * time.Hour,
	})
	conversationSvc := convApp.NewService(convApp.ServiceConfig{
		ConvRepo: conversationRepo,
		MsgRepo:  messageRepo,
	})

	// Initialize handlers
	whatsappHdlr := whatsappHandler.NewHandler(whatsappHandler.HandlerConfig{
		WhatsAppSvc:        whatsappSvc,
		ConversationSvc:    conversationSvc,
		DocumentSvc:        documentSvc,
		WebhookVerifyToken: cfg.WhatsApp.WebhookVerifyToken,
		Log:                log,
	})
	documentHdlr := documentHandler.NewHandler(documentSvc, log)
	ragHdlr := ragHandler.NewHandler(documentSvc, log)
	authHdlr := authHandler.NewHandler(userSvc, log)
	conversationHdlr := conversationHandler.NewHandler(conversationSvc, log)

	// Initialize auth middleware
	authMw := middleware.AuthMiddleware(userSvc)
	adminMw := middleware.RequireRole("admin")

	// Initialize rate limiter (100 requests per minute per IP)
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)

	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(requestIDMiddleware())
	engine.Use(loggingMiddleware(log))
	engine.Use(corsMiddleware())
	engine.Use(middleware.RateLimit(rateLimiter))

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	engine.GET("/readyz", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		if err := dbClient.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "error",
				"database": "disconnected",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
		})
	})

	v1 := engine.Group("/api/v1")

	// Public routes - auth endpoints
	authHandler.Register(v1, authHdlr, authMw)

	// Public routes - WhatsApp webhooks (uses its own verification)
	whatsappHandler.Register(v1, whatsappHdlr)

	// Protected routes - RAG query (authenticated users)
	ragGroup := v1.Group("/rag")
	ragGroup.Use(authMw)
	ragHandler.Register(ragGroup, ragHdlr)

	// Protected routes - Document management (admin only)
	docGroup := v1.Group("/documents")
	docGroup.Use(authMw, adminMw)
	documentHandler.Register(docGroup, documentHdlr)

	// Protected routes - Conversation history (admin only)
	convGroup := v1.Group("/conversations")
	convGroup.Use(authMw, adminMw)
	conversationHandler.Register(convGroup, conversationHdlr)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Info("Server listening", "address", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	if err := dbClient.Close(shutdownCtx); err != nil {
		log.Error("Failed to close database connection", "error", err)
	}

	log.Info("Server stopped")
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		c.Next()
	}
}

func loggingMiddleware(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		log.Info("request completed",
			"request_id", c.GetString("request_id"),
			"method", c.Request.Method,
			"path", path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
