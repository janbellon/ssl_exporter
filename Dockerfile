# Build
ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG NO_PROXY

FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o ssl_exporter ./cmd/ssl_exporter

# Runtime
FROM alpine:latest

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/ssl_exporter .

EXPOSE 9123

CMD ["./ssl_exporter"]