# Vertex Middleware

A lightweight Go middleware that exposes an OpenAI-compatible REST API and proxies requests to Google Cloud Vertex AI (Gemini models). Useful for platforms that support OpenAI's API format but not Vertex AI's native format or authentication.

```
Client (OpenAI API format)
    → Vertex Middleware  (auth, request/response translation)
        → Google Cloud Vertex AI (Gemini)
```

## Features

- OpenAI-compatible `/v1/chat/completions` and `/v1/models` endpoints
- Streaming support (Server-Sent Events)
- Multiple API keys, each mapped to its own GCP project and credentials
- Single binary, minimal dependencies

## Prerequisites

- Go 1.21+
- A Google Cloud project with the Vertex AI API enabled
- A GCP service account with the **Vertex AI User** role, and its JSON key downloaded

## Setup

### 1. Get a service account key

In the GCP Console, go to **IAM & Admin → Service Accounts**, create a service account, grant it the `Vertex AI User` role, then create and download a JSON key. Place it somewhere accessible, e.g. `credentials/my-project.json`.

### 2. Configure

Copy the sample config and edit it:

```bash
cp config.yaml.sample config.yaml
```

```yaml
server:
  port: "8080"

api_keys:
  "sk-your-generated-key":
    credential_file: "credentials/my-project.json"
    project_id: "my-gcp-project-id"
    location: "us-central1"       # or "global" for the global endpoint
    default_model: "gemini-2.5-flash"
```

Each entry under `api_keys` is an independent API key that clients will use. You can have as many as you need, each pointing to a different GCP project, region, or credential file.

**Supported locations:** any Vertex AI region (`us-central1`, `europe-west4`, etc.) or `global`.

### 3. Generate an API key

```bash
go run . generate-key
# sk-3f2a1b...
```

Paste the output as a key in `config.yaml`.

### 4. Run

```bash
go run .
# or build a binary:
go build -o vertex-middleware .
./vertex-middleware
# optionally pass a config path:
./vertex-middleware /path/to/config.yaml
```

## Docker

```bash
docker build -t vertex-middleware .
docker run -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  -v $(pwd)/credentials:/app/credentials \
  vertex-middleware
```

## API

### Authentication

All `/v1` endpoints require a Bearer token matching a key in `config.yaml`:

```
Authorization: Bearer sk-your-generated-key
```

### POST /v1/chat/completions

Follows the [OpenAI Chat Completions API](https://platform.openai.com/docs/api-reference/chat).

**Request:**
```json
{
  "model": "gemini-2.5-flash",
  "messages": [
    { "role": "system", "content": "You are a helpful assistant." },
    { "role": "user",   "content": "Hello!" }
  ],
  "temperature": 0.7,
  "max_tokens": 1024,
  "stream": false
}
```

If `model` is omitted, the `default_model` from the matching key's config is used.

**Response:**
```json
{
  "id": "chatcmpl-...",
  "object": "chat.completion",
  "created": 1234567890,
  "model": "gemini-2.5-flash",
  "choices": [
    {
      "index": 0,
      "message": { "role": "assistant", "content": "Hello! How can I help?" },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 12,
    "completion_tokens": 9,
    "total_tokens": 21
  }
}
```

Set `"stream": true` to receive a standard SSE stream of `chat.completion.chunk` events.

### GET /v1/models

Returns a list of known Gemini models in OpenAI format.

### GET /health

Unauthenticated health check. Returns `{"status": "ok"}`.

## Testing with curl

```bash
# Non-streaming
curl -s -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-your-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "Say hello"}]
  }'

# Streaming
curl -s -X POST http://localhost:8080/v1/chat/completions \
  -H "Authorization: Bearer sk-your-key" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "Say hello"}],
    "stream": true
  }'

# List models
curl -s http://localhost:8080/v1/models \
  -H "Authorization: Bearer sk-your-key"
```

