FROM golang:1.24.2-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY cmd ./cmd

RUN go build -o /bin/senju-api ./cmd/api

FROM alpine:3.21

RUN apk add --no-cache ca-certificates wget

WORKDIR /srv
COPY --from=builder /bin/senju-api /usr/local/bin/senju-api

EXPOSE 8080

ENTRYPOINT ["senju-api"]
