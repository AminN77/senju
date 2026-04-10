FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal
COPY openapi ./openapi

ARG VERSION=0.1.0
ARG COMMIT=unknown
ARG BUILD_TIME=unknown

RUN CGO_ENABLED=0 go build -trimpath -ldflags "-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}" -o /bin/senju-api ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache ca-certificates wget

WORKDIR /srv
COPY --from=builder /bin/senju-api /usr/local/bin/senju-api

EXPOSE 8080

ENTRYPOINT ["senju-api"]
