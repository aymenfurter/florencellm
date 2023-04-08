# Use the official Go image as a base
FROM golang:1.20

# Set the working directory in the container
WORKDIR /app

# Copy the Go module files into the container
COPY go.mod .
COPY go.sum .

# Download the Go module dependencies
RUN go mod download

# Copy the rest of the source code into the container
COPY . .

# Build the Go binary
RUN go build -o repository-microservice

# Expose the microservice port
EXPOSE 8081

# Set the entrypoint for the container
ENTRYPOINT ["/app/repository-microservice"]

