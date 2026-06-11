FROM golang:1.26-bullseye

# Install Git and clean up package lists to keep the image lean
RUN apt-get update && apt-get install -y \
    git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files first to leverage Docker caching layers
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY *.go ./

# Compile the core management orchestrator
RUN go build -o hs-mcp-manager-core-backend .

# Create the workspace directory inside the container
RUN mkdir /workspace

# Expose the application port
EXPOSE 8080

# Run the manager binary
CMD ["./hs-mcp-manager-core-backend"]