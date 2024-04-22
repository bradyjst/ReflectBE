# Start from the official Go image.
FROM golang:1.18-alpine

# Install general dependencies
RUN apk add --no-cache git

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Install Air for live reloading
RUN go install github.com/cosmtrek/air@latest

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Copy Air configuration file
COPY .air.toml ./

# Command to run the executable through Air
CMD ["air"]
