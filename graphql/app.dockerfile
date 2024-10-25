# Base image for building our Go application, using Golang 1.23.2 on Alpine 3.19
FROM golang:1.23.2-alpine3.19 AS build

# Install necessary dependencies (gcc, g++, make, ca-certificates) for building Go app
RUN apk --no-cache add \
    gcc \        
    g++ \       
    make \      
    ca-certificates

# Set working directory to /go/src/github.com/pirateunclejack/go-grpc-graphql-microservice
WORKDIR /go/src/github.com/pirateunclejack/go-grpc-graphql-microservice

# Copy necessary files (go.mod and go.sum) into the build context
COPY go.mod go.sum ./

COPY vendor vendor
COPY catalog catalog
COPY account account
COPY order order
COPY graphql graphql

# Build our Go application using GO111MODULE=on, with dependencies from vendor directory
RUN GO111MODULE=on go build -mod vendor -o /go/bin/app ./graphql


# New base image for running our application, using Alpine 3.19
FROM alpine:3.19

# Set working directory to /usr/bin
WORKDIR /usr/bin

# Copy the built executable from the build stage into this stage
COPY --from=build /go/bin/ .

# Expose port 8080 for our application to listen on
EXPOSE 8080

# Set default command to run when container is started
CMD ["app"]
