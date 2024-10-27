# Use the official Golang image as the base image
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o consumer ./main.go

# Use a minimal base image to reduce the size of the final image
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app/

# Copy the binary from the builder stage
COPY --from=builder /app/consumer .

# Command to run the executable
CMD ["./consumer"]