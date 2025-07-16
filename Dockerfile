# Stage 1: The Build Stage
# Uses a full Go SDK image to compile the application
FROM golang:1.24-alpine AS builder

# Set the current working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker layer caching.
# This ensures these layers are only rebuilt if dependencies change.
COPY go.mod ./
COPY go.sum ./

# Download Go modules (dependencies)
RUN go mod download

# Copy the rest of your application source code
COPY . .

# Build the Go application
# -o output_name: Specifies the output executable name
# -ldflags "-s -w": Reduces binary size by stripping debug info and symbol tables
# ./cmd/my-app: Path to your main package (adjust as needed)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o my-app main.go

# Stage 2: The Final (Runtime) Stage
# Uses a minimal base image to run the compiled application
FROM alpine:latest AS final

# Optional: Install ca-certificates if your app makes HTTPS calls
# Alpine uses musl libc, so ca-certificates might be needed for TLS
RUN apk add --no-cache ca-certificates

# Set the current working directory for the final application
WORKDIR /app

# Copy only the compiled executable from the 'builder' stage
COPY --from=builder /app/my-app .

# Expose the port your application listens on (if any)
EXPOSE 8080

# Command to run the application when the container starts
CMD ["./my-app"]