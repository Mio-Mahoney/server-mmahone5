FROM public.ecr.aws/docker/library/golang:latest AS build

# Copy source
WORKDIR /build
COPY . .

# Download dependencies
RUN go mod download

# Build the application for Linux without CGO dependencies
RUN CGO_ENABLED=0 GOOS=linux go build -a -o main .

# Use Alpine to extract CA certificates
FROM alpine:latest AS certs
RUN apk --no-cache add ca-certificates

# Final stage using scratch
FROM scratch
# Copy CA certificates from the certs stage
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /app

# Copy the built binary from the build stage
COPY --from=build /build/main ./

# Identify listening port
EXPOSE 8080

CMD ["./main"]

