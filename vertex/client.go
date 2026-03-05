package vertex

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2/google"
	"vertexmiddleware/config"
)

const cloudPlatformScope = "https://www.googleapis.com/auth/cloud-platform"

type tokenProvider interface {
	Token() (string, error)
}

type Client struct {
	keyConfig  config.KeyConfig
	httpClient *http.Client
	tokens     tokenProvider
}

func NewClient(keyConfig config.KeyConfig) (*Client, error) {
	ctx := context.Background()

	data, err := os.ReadFile(keyConfig.CredentialFile)
	if err != nil {
		return nil, fmt.Errorf("reading credentials file %q: %w", keyConfig.CredentialFile, err)
	}

	creds, err := google.CredentialsFromJSON(ctx, data, cloudPlatformScope)
	if err != nil {
		return nil, fmt.Errorf("parsing credentials: %w", err)
	}

	return &Client{
		keyConfig:  keyConfig,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		tokens:     &oauth2TokenSource{ts: creds.TokenSource},
	}, nil
}

// GenerateContent calls the Vertex AI generateContent endpoint (non-streaming).
func (c *Client) GenerateContent(ctx context.Context, model string, req GenerateContentRequest) (*GenerateContentResponse, error) {
	token, err := c.tokens.Token()
	if err != nil {
		return nil, fmt.Errorf("obtaining access token: %w", err)
	}

	url := c.endpointURL(model, "generateContent")

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Vertex AI: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vertex API returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result GenerateContentResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	return &result, nil
}

// StreamGenerateContent calls the Vertex AI streamGenerateContent endpoint.
// The caller is responsible for closing the returned ReadCloser.
func (c *Client) StreamGenerateContent(ctx context.Context, model string, req GenerateContentRequest) (io.ReadCloser, error) {
	token, err := c.tokens.Token()
	if err != nil {
		return nil, fmt.Errorf("obtaining access token: %w", err)
	}

	// ?alt=sse requests SSE-formatted streaming response
	url := c.endpointURL(model, "streamGenerateContent") + "?alt=sse"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")

	// No timeout for streaming — rely on context cancellation from the HTTP handler.
	resp, err := (&http.Client{}).Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Vertex AI stream: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("vertex API returned %d: %s", resp.StatusCode, string(errBody))
	}

	return resp.Body, nil
}

func (c *Client) endpointURL(model, method string) string {
	return fmt.Sprintf(
		"https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:%s",
		c.keyConfig.Location,
		c.keyConfig.ProjectID,
		c.keyConfig.Location,
		model,
		method,
	)
}
