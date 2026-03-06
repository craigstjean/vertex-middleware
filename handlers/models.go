package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/craigstjean/vertex-middleware/types"
)

var knownModels = []types.Model{
	{ID: "gemini-2.0-flash-001", Object: "model", Created: time.Now().Unix(), OwnedBy: "google"},
	{ID: "gemini-2.0-flash-lite-001", Object: "model", Created: time.Now().Unix(), OwnedBy: "google"},
	{ID: "gemini-2.5-flash-lite", Object: "model", Created: time.Now().Unix(), OwnedBy: "google"},
	{ID: "gemini-2.5-flash", Object: "model", Created: time.Now().Unix(), OwnedBy: "google"},
	{ID: "gemini-2.5-pro", Object: "model", Created: time.Now().Unix(), OwnedBy: "google"},
	{ID: "gemini-3.1-pro-preview", Object: "model", Created: time.Now().Unix(), OwnedBy: "google"},
	{ID: "gemini-3-flash-preview", Object: "model", Created: time.Now().Unix(), OwnedBy: "google"},
	{ID: "gemini-3.1-flash-lite-preview", Object: "model", Created: time.Now().Unix(), OwnedBy: "google"},
}

func ListModels() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, types.ModelList{
			Object: "list",
			Data:   knownModels,
		})
	}
}
