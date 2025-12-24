package main

import (
	"context"

	"log"

	userApp "github.com/elprogramadorgt/lucidRAG/internal/application/user"
	"github.com/elprogramadorgt/lucidRAG/internal/repository/mongo"
	userV1 "github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/user"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logger

	gin.ForceConsoleColor()
	engine := gin.Default()
	engine.ContextWithFallback = true
	ctx := context.Background()

	client, err := mongo.NewClient(ctx, "mongodb://root:lucidrag@localhost:27019", "lucid")
	if err != nil {
		log.Fatal(err)
	}

	// whatsappRepo := mongo.NewWhatsappRepo(client)
	userRepo := mongo.NewUserRepo(client)

	// Load configuration
	// cfg, err := config.Load()
	// if err != nil {
	// 	logrus.Error(err)
	// 	os.Exit(1)
	// }

	svcUser := userApp.NewService(userRepo)
	ctlUser := userV1.NewHandler(svcUser)
	v1 := engine.Group("/v1")

	userV1.Register(v1, ctlUser)

	engine.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	log.Fatal(engine.Run("0.0.0.0:8080"))

}
