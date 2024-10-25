FROM public.ecr.aws/docker/library/golang:latest AS build

# Copy source
WORKDIR /build
COPY . .

# Download dependencies
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -a -o main .

FROM scratch

# Add support for HTTPS and time zones
#RUN apk update && \
#    apk upgrade && \
#    apk add ca-certificates && \
#    apk add tzdata

WORKDIR /app

COPY --from=build /build/main ./

# RUN pwd && find .

# Identify listening port
EXPOSE 8080

CMD ["./main"]
