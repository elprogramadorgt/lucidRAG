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
	ragApp "github.com/elprogramadorgt/lucidRAG/internal/application/rag"
	userApp "github.com/elprogramadorgt/lucidRAG/internal/application/user"
	"github.com/elprogramadorgt/lucidRAG/internal/application/whatsapp"
	"github.com/elprogramadorgt/lucidRAG/internal/config"
	"github.com/elprogramadorgt/lucidRAG/internal/repository/mongo"
	"github.com/elprogramadorgt/lucidRAG/internal/transport/http/middleware"
	authHandler "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/auth"
	conversationHandler "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/conversation"
	documentHandler "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/document"
	healthHandler "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/health"
	ragHandler "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/rag"
	systemHandler "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/system"
	whatsappHandler "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/whatsapp"
	"github.com/elprogramadorgt/lucidRAG/pkg/chunker"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/elprogramadorgt/lucidRAG/pkg/openai"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var version = "dev" // set via -ldflags at build time

func main() {
	startTime := time.Now()
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		os.Exit(1)
	}

	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx := context.Background()
	db, err := mongo.NewClient(ctx, cfg.Database.MongoURI(), cfg.Database.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "mongo: %v\n", err)
		os.Exit(1)
	}

	logRepo := mongo.NewLogRepo(db)
	if err := logRepo.EnsureIndexes(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "log indexes: %v\n", err)
		os.Exit(1)
	}
	log := logger.New(logger.Options{
		Level: cfg.Server.LogLevel(),
		JSON:  cfg.Server.Environment == "production",
		Store: logRepo,
	})

	var openaiClient *openai.Client
	if cfg.RAG.OpenAIAPIKey != "" {
		openaiClient = openai.NewClient(cfg.RAG.OpenAIAPIKey)
	}

	chunkRepo := mongo.NewChunkRepo(db)
	ragSvc := ragApp.NewService(ragApp.ServiceConfig{
		ChunkRepo:      chunkRepo,
		OpenAIClient:   openaiClient,
		Chunker:        chunker.New(cfg.RAG.ChunkSize, cfg.RAG.ChunkOverlap),
		EmbeddingModel: cfg.RAG.EmbeddingModel,
		ModelName:      cfg.RAG.ModelName,
		Log:            log,
	})

	whatsappSvc := whatsapp.NewService(mongo.NewWhatsappRepo(db))
	documentSvc := docApp.NewService(docApp.ServiceConfig{
		Repo:   mongo.NewDocumentRepo(db),
		RAGSvc: ragSvc,
		Log:    log,
	})
	userSvc := userApp.NewService(userApp.ServiceConfig{
		Repo: mongo.NewUserRepo(db), JWTSecret: cfg.Auth.JWTSecret,
		JWTExpiry: time.Duration(cfg.Auth.JWTExpiryHours) * time.Hour,
	})
	conversationSvc := convApp.NewService(convApp.ServiceConfig{
		ConvRepo: mongo.NewConversationRepo(db), MsgRepo: mongo.NewMessageRepo(db),
	})

	whatsappHdlr := whatsappHandler.NewHandler(whatsappHandler.HandlerConfig{
		WhatsAppSvc: whatsappSvc, ConversationSvc: conversationSvc, RAGSvc: ragSvc,
		WebhookVerifyToken: cfg.WhatsApp.WebhookVerifyToken, Log: log,
	})

	authMw, adminMw := middleware.AuthMiddleware(userSvc), middleware.RequireRole("admin")
	rateLimiter := middleware.NewRateLimiter(100, time.Minute)

	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestID(), middleware.Logger(log))
	r.Use(middleware.CORS([]string{"http://localhost:4200", "http://localhost:8080"}))
	r.Use(middleware.RateLimit(rateLimiter))

	healthHandler.Register(r, healthHandler.NewHandler(db))

	v1 := r.Group("/api/v1")
	cookieCfg := authHandler.CookieConfig{
		Domain:      cfg.Auth.CookieDomain,
		Secure:      cfg.Auth.CookieSecure,
		ExpiryHours: cfg.Auth.JWTExpiryHours,
	}
	authHandler.Register(v1, authHandler.NewHandler(userSvc, log, cookieCfg), authMw)
	authHandler.RegisterOAuth(v1, authHandler.NewOAuthHandler(userSvc, log, cfg.Auth.OAuth, cookieCfg))
	whatsappHandler.Register(v1, whatsappHdlr)
	ragHandler.Register(v1.Group("/rag", authMw), ragHandler.NewHandler(ragSvc, log))
	documentHandler.Register(v1.Group("/documents", authMw), documentHandler.NewHandler(documentSvc, log))
	conversationHandler.Register(v1.Group("/conversations", authMw), conversationHandler.NewHandler(conversationSvc, log))
	systemHandler.Register(v1.Group("/system", authMw, adminMw), systemHandler.NewHandler(systemHandler.HandlerConfig{
		Repo:        logRepo,
		DB:          db,
		Log:         log,
		StartTime:   startTime,
		Environment: cfg.Server.Environment,
		Version:     version,
	}))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{Addr: addr, Handler: r, ReadTimeout: 15 * time.Second, WriteTimeout: 15 * time.Second}

	go func() {
		log.Info("listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	rateLimiter.Stop()
	log.Stop()
	_ = db.Close(shutdownCtx)
}
