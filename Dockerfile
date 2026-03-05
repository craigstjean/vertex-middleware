FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o vertex-middleware .

# ---

FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/vertex-middleware .

EXPOSE 8080

ENTRYPOINT ["./vertex-middleware"]
CMD ["config.yaml"]
