# Use golang:1.23.2-alpine3.19 as our base image, and give it an alias of "build"
FROM golang:1.23.2-alpine3.19 AS build

# Install necessary dependencies (gcc, g++, make, and ca-certificates) in the build stage
RUN apk --no-cache add gcc g++ make ca-certificates

# Change directory to the Go source code location
WORKDIR /go/src/github.com/pirateunclejack/go-grpc-graphql-microservice

# Copy our Go module file (go.mod), sum file (go.sum), and vendor directory into the build context
COPY go.mod go.sum ./

# Copy our vendor directory into the build context
COPY vendor vendor

# Copy our account code into the build context
COPY account account

# Build our Go application using the GO111MODULE=on flag, with the output file named "app"
RUN GO111MODULE=on go build -mod vendor -o /go/bin/app ./account/cmd/account

# Use a new base image (alpine:3.19) and give it an alias of "runtime"
FROM alpine:3.19

# Change directory to the runtime context
WORKDIR /usr/bin

# Copy our built Go application from the build stage into the runtime context
COPY --from=build /go/bin/ .

# Expose port 8080 for external access
EXPOSE 8080

# Set the default command to run when the container is started (our Go application)
CMD ["app"]
