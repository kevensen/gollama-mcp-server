FROM golang:1.24-alpine3.21 AS builder

WORKDIR /app

# pre-copy/cache go.mod for pre-downloading dependencies and only redownloading them in subsequent builds if they change
COPY go.mod go.sum ./
RUN go mod download

COPY . .
# Build the application (CGO disabled since we use custom HTTP-based ChromaDB client)
RUN CGO_ENABLED=0 GOOS=linux go build -a \
  -ldflags '-w -s' \
  -o golang-mcp-server ./cmd/main.go

FROM golang:1.24-alpine3.21
EXPOSE 8080
RUN addgroup -g 1001 appgroup && \
  adduser -u 1001 -G appgroup -s /bin/sh -D appuser

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/golang-mcp-server .

# Set ownership of the app directory
RUN chown -R appuser:appgroup /app

CMD ["/app/golang-mcp-server", "-port", "8080"]