# Stage 1: Build lightweight Golang binary
FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/pos-api ./cmd/api

# Stage 2: Final minimal runtime image
FROM alpine:3.19

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/pos-api /app/pos-api

EXPOSE 3000

ENTRYPOINT ["/app/pos-api"]
