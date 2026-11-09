# Build
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

ARG TARGETOS
ARG TARGETARCH
ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG NO_PROXY

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build \
    -trimpath \
    -ldflags="-s -w" \
    -o ssl_exporter .

# Runtime
FROM alpine:3.22

RUN apk add --no-cache ca-certificates \
    && addgroup -S ssl_exporter \
    && adduser -S -G ssl_exporter -H -s /sbin/nologin ssl_exporter

WORKDIR /app

COPY --from=builder /app/ssl_exporter .

USER ssl_exporter

EXPOSE 9123

CMD ["./ssl_exporter"]
