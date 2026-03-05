package vertex

import (
	"vertexmiddleware/types"
)

// ToVertexRequest converts an OpenAI ChatCompletionRequest to a Vertex GenerateContentRequest.
func ToVertexRequest(req types.ChatCompletionRequest) GenerateContentRequest {
	var contents []Content
	var systemInstruction *Content

	for _, msg := range req.Messages {
		switch msg.Role {
		case "system":
			// Vertex handles system prompts separately via systemInstruction.
			systemInstruction = &Content{
				Parts: []Part{{Text: msg.Content}},
			}
		case "assistant":
			contents = append(contents, Content{
				Role:  "model", // Vertex uses "model" where OpenAI uses "assistant"
				Parts: []Part{{Text: msg.Content}},
			})
		default: // "user"
			contents = append(contents, Content{
				Role:  msg.Role,
				Parts: []Part{{Text: msg.Content}},
			})
		}
	}

	genConfig := GenerationConfig{
		Temperature:     req.Temperature,
		MaxOutputTokens: req.MaxTokens,
		TopP:            req.TopP,
	}
	if len(req.Stop) > 0 {
		genConfig.StopSequences = req.Stop
	}
	if req.N != nil {
		genConfig.CandidateCount = req.N
	}

	return GenerateContentRequest{
		Contents:          contents,
		SystemInstruction: systemInstruction,
		GenerationConfig:  genConfig,
	}
}

// FromVertexResponse converts a Vertex GenerateContentResponse to an OpenAI ChatCompletionResponse.
func FromVertexResponse(vertexResp *GenerateContentResponse, model, id string, created int64) types.ChatCompletionResponse {
	choices := make([]types.Choice, 0, len(vertexResp.Candidates))

	for _, candidate := range vertexResp.Candidates {
		text := extractText(candidate.Content)
		choices = append(choices, types.Choice{
			Index: candidate.Index,
			Message: types.Message{
				Role:    "assistant",
				Content: text,
			},
			FinishReason: mapFinishReason(candidate.FinishReason),
		})
	}

	return types.ChatCompletionResponse{
		ID:      id,
		Object:  "chat.completion",
		Created: created,
		Model:   model,
		Choices: choices,
		Usage: types.Usage{
			PromptTokens:     vertexResp.UsageMetadata.PromptTokenCount,
			CompletionTokens: vertexResp.UsageMetadata.CandidatesTokenCount,
			TotalTokens:      vertexResp.UsageMetadata.TotalTokenCount,
		},
	}
}

// extractText concatenates all text parts from a Content block.
func extractText(content Content) string {
	var result string
	for _, part := range content.Parts {
		result += part.Text
	}
	return result
}

// MapFinishReason converts Vertex finish reasons to OpenAI finish reasons.
// Exported so handlers can reuse it for streaming chunks.
func MapFinishReason(vertexReason string) string {
	switch vertexReason {
	case "STOP":
		return "stop"
	case "MAX_TOKENS":
		return "length"
	case "SAFETY", "RECITATION":
		return "content_filter"
	default:
		return ""
	}
}

func mapFinishReason(vertexReason string) string {
	r := MapFinishReason(vertexReason)
	if r == "" {
		return "stop"
	}
	return r
}
