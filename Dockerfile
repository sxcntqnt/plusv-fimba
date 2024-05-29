# Start with Golang as base image
FROM golang:latest as builder
# Set environment variables
ENV GO111MODULE=on

# Copy the source code into the container
COPY ../fimba /sxcntcnquntns/
 
# Set working directory
WORKDIR /sxcntcnquntns/

# Download Go modules
RUN go mod download

# Build the server binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/server main.go

# Second stage, use scratch as minimal base image
FROM scratch

# Copy the built binary from the builder stage
COPY --from=builder /bin/server /bin/server

# Expose ports
EXPOSE 3420 3420

# Command to run the server
CMD ["./bin/server"]
