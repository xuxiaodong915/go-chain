# Stage 1: Build
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o recipe-server .

# Stage 2: Run
FROM alpine:3.19
WORKDIR /app

# Runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Copy binary
COPY --from=builder /app/recipe-server .

# Create necessary directories
RUN mkdir -p data uploads

# Copy seed images (sichuan_1.jpg, etc.) for initial image assignment
COPY --from=builder /app/images ./images

# Expose API port
EXPOSE 8080

# Use a non-root user for security
RUN adduser -D appuser && chown -R appuser:appuser /app
USER appuser

# Start
CMD ["./recipe-server"]
