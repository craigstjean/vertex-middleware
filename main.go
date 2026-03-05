package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/craigstjean/vertexmiddleware/config"
	"github.com/craigstjean/vertexmiddleware/handlers"
	"github.com/craigstjean/vertexmiddleware/middleware"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "generate-key" {
		key, err := generateKey()
		if err != nil {
			log.Fatalf("Failed to generate key: %v", err)
		}
		fmt.Println(key)
		return
	}

	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	r := gin.Default()

	// Health check — no auth required.
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/v1")
	v1.Use(middleware.APIKeyAuth(cfg))
	{
		v1.POST("/chat/completions", handlers.ChatCompletions(cfg))
		v1.GET("/models", handlers.ListModels())
	}

	log.Printf("Vertex AI middleware listening on :%s", cfg.Server.Port)
	if err := r.Run(":" + cfg.Server.Port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// generateKey returns a cryptographically random API key in the form "sk-<32 hex bytes>".
func generateKey() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "sk-" + hex.EncodeToString(b), nil
}
