FROM golang:1.21-alpine

# Install git and git-lfs
RUN apk add --no-cache git git-lfs

# Set the working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o /harness-migrator

# Set the entrypoint
ENTRYPOINT ["/harness-migrator"]
