# Start from the official Go image.
FROM golang:1.22-alpine

# Install necessary packages
RUN apk update && apk add --no-cache git bash curl

# Install migrate CLI
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.15.1/migrate.linux-amd64.tar.gz | tar xz -C /usr/local/bin

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Install Air for live reloading
RUN go install github.com/air-verse/air

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Copy Air configuration file
COPY .air.toml ./

# Command to run the executable through Air
CMD ["air"]
