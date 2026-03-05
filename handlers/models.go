package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"vertexmiddleware/types"
)

var knownModels = []types.Model{
	{ID: "gemini-2.0-flash-001", Object: "model", Created: time.Now().Unix(), OwnedBy: "google"},
	{ID: "gemini-2.0-flash-lite-001", Object: "model", Created: time.Now().Unix(), OwnedBy: "google"},
	{ID: "gemini-1.5-pro-002", Object: "model", Created: time.Now().Unix(), OwnedBy: "google"},
	{ID: "gemini-1.5-flash-002", Object: "model", Created: time.Now().Unix(), OwnedBy: "google"},
	{ID: "gemini-1.0-pro-002", Object: "model", Created: time.Now().Unix(), OwnedBy: "google"},
}

func ListModels() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, types.ModelList{
			Object: "list",
			Data:   knownModels,
		})
	}
}
