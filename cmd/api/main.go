package main

import (
	"context"
	"os"

	"log"

	"github.com/elprogramadorgt/lucidRAG/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	// Initialize logger

	gin.ForceConsoleColor()
	engine := gin.Default()
	engine.ContextWithFallback = true
	ctx := context.Background()

	client, err := mongo.NewClient(ctx, "mongodb://webapp:eduquest@localhost:27018/eduquestdb", "eduquestdb")
	if err != nil {
		log.Fatal(err)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.Error(err)
		os.Exit(1)
	}

	v1 := engine.Group("/v1")

	engine.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	log.Fatal(engine.Run("0.0.0.0:8080"))

}
