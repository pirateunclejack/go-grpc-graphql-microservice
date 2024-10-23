# Base image for building the Go application
FROM golang:1.23.2-alpine3.19 AS build

# Install necessary packages (gcc, make, ca-certificates) on Alpine 3.19
RUN apk --no-cache add gcc g++ make ca-certificates

# Set working directory to /go/src/github.com/pirateunclejack/go-grpc-graphql-microservice
WORKDIR /go/src/github.com/pirateunclejack/go-grpc-graphql-microservice

# Copy Go module file and sum into the build context
COPY go.mod go.sum ./

# Copy vendor directory into the build context (if using a vendor directory)
COPY vendor vendor

# Copy individual files (account, catalog, order) into the build context
COPY account account
COPY catalog catalog
COPY order order

# Build the Go application with GO111MODULE=on and mod vendor
RUN GO111MODULE=on go build -mod vendor -o /go/bin/app ./order/cmd/order

# New base image for running the application (Alpine 3.19)
FROM alpine:3.19

# Set working directory to /usr/bin
WORKDIR /usr/bin

# Copy the built Go application from the previous step into this new context
COPY --from=build /go/bin/ .

# Expose port 8080 for the application to listen on
EXPOSE 8080

# Define the default command to run when the container starts (the "app" executable)
CMD ["app"]
