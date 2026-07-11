# --- Stage 1: Build the Go Binary ---
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy dependency definition files first (helps with Docker caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Compile the Go application to a static binary
# CGO_ENABLED=0 builds a standalone binary with no external C dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -o shortener-server cmd/server/main.go

# --- Stage 2: Create the lightweight Run container ---
FROM alpine:latest

# Install certificates (needed if your app makes outbound HTTPS requests)
RUN apk --no-cache add ca-certificates

WORKDIR /root/
# Copy only the compiled binary from the builder stage
COPY --from=builder /app/shortener-server .

# Expose port 8080 to the outside world
EXPOSE 8080
 
# Command to run when the container starts
CMD ["./shortener-server"]
