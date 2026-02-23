# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY services/backend/go.mod .
COPY services/backend/go.sum* .
RUN go mod download || true
COPY services/backend/main.go .

RUN go build -o backend main.go

# Final stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/backend .

EXPOSE 8081

CMD ["./backend"]
