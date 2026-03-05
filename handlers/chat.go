package handlers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/craigstjean/vertexmiddleware/config"
	"github.com/craigstjean/vertexmiddleware/middleware"
	"github.com/craigstjean/vertexmiddleware/types"
	"github.com/craigstjean/vertexmiddleware/vertex"
)

// clientCache avoids re-initialising a Vertex client (and its token source) on every request.
var clientCache = newSafeClientCache()

func ChatCompletions(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req types.ChatCompletionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: types.ErrorDetail{
					Message: "Invalid request body: " + err.Error(),
					Type:    "invalid_request_error",
				},
			})
			return
		}

		if len(req.Messages) == 0 {
			c.JSON(http.StatusBadRequest, types.ErrorResponse{
				Error: types.ErrorDetail{
					Message: "messages must not be empty",
					Type:    "invalid_request_error",
				},
			})
			return
		}

		keyConfig := c.MustGet(middleware.KeyConfigCtxKey).(config.KeyConfig)

		client, err := clientCache.get(keyConfig)
		if err != nil {
			c.JSON(http.StatusInternalServerError, types.ErrorResponse{
				Error: types.ErrorDetail{
					Message: "Failed to initialise Vertex client: " + err.Error(),
					Type:    "server_error",
				},
			})
			return
		}

		model := req.Model
		if model == "" {
			model = keyConfig.DefaultModel
		}
		if model == "" {
			model = "gemini-1.5-pro-002"
		}

		vertexReq := vertex.ToVertexRequest(req)

		if req.Stream {
			handleStream(c, client, model, req.Model, vertexReq)
			return
		}

		vertexResp, err := client.GenerateContent(c.Request.Context(), model, vertexReq)
		if err != nil {
			c.JSON(http.StatusBadGateway, types.ErrorResponse{
				Error: types.ErrorDetail{
					Message: err.Error(),
					Type:    "server_error",
				},
			})
			return
		}

		id := newID()
		created := time.Now().Unix()
		c.JSON(http.StatusOK, vertex.FromVertexResponse(vertexResp, req.Model, id, created))
	}
}

func handleStream(c *gin.Context, client *vertex.Client, model, originalModel string, vertexReq vertex.GenerateContentRequest) {
	stream, err := client.StreamGenerateContent(c.Request.Context(), model, vertexReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, types.ErrorResponse{
			Error: types.ErrorDetail{
				Message: err.Error(),
				Type:    "server_error",
			},
		})
		return
	}
	defer stream.Close()

	id := newID()
	created := time.Now().Unix()

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // Disable nginx buffering

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, types.ErrorResponse{
			Error: types.ErrorDetail{Message: "Streaming not supported by server", Type: "server_error"},
		})
		return
	}

	firstChunk := true
	scanner := bufio.NewScanner(stream)

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var vertexChunk vertex.GenerateContentResponse
		if err := json.Unmarshal([]byte(data), &vertexChunk); err != nil {
			continue
		}

		for _, candidate := range vertexChunk.Candidates {
			text := ""
			for _, part := range candidate.Content.Parts {
				text += part.Text
			}

			delta := types.Delta{Content: text}
			if firstChunk {
				delta.Role = "assistant"
				firstChunk = false
			}

			finishReasonStr := vertex.MapFinishReason(candidate.FinishReason)
			var finishReason *string
			if finishReasonStr != "" {
				finishReason = &finishReasonStr
			}

			chunk := types.ChatCompletionStreamResponse{
				ID:      id,
				Object:  "chat.completion.chunk",
				Created: created,
				Model:   originalModel,
				Choices: []types.StreamChoice{
					{
						Index:        candidate.Index,
						Delta:        delta,
						FinishReason: finishReason,
					},
				},
			}

			chunkJSON, err := json.Marshal(chunk)
			if err != nil {
				continue
			}

			fmt.Fprintf(c.Writer, "data: %s\n\n", chunkJSON)
			flusher.Flush()
		}
	}

	fmt.Fprint(c.Writer, "data: [DONE]\n\n")
	flusher.Flush()
}

func newID() string {
	return fmt.Sprintf("chatcmpl-%016x", rand.Int63())
}
