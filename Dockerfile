FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY vendor/ vendor/
COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -ldflags="-s -w" -o /bitrise-cache .

FROM alpine:3.19

RUN apk add --no-cache tar zstd ca-certificates curl

# Install envman
RUN curl -fL "https://github.com/bitrise-io/envman/releases/latest/download/envman-$(uname -s)-$(uname -m)" \
    -o /usr/local/bin/envman \
    && chmod +x /usr/local/bin/envman

COPY --from=builder /bitrise-cache /bitrise-cache
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
