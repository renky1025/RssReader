# Build backend
FROM --platform=linux/amd64 golang:1.21-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o /app/server ./cmd/server

# Final image
FROM --platform=linux/amd64 alpine:3.19
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

# Copy backend binary
COPY --from=builder /app/server /app/server

# Copy migrations
COPY migrations /app/migrations

# Copy data files
COPY data /app/data

# Create data directory for database
RUN mkdir -p /app/data

EXPOSE 8080

ENV DATABASE_PATH=/app/data/rssreader.db
ENV SERVER_PORT=8080

CMD ["/app/server"]
