# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY services/frontend/go.mod .
COPY services/frontend/main.go .

RUN go build -o frontend main.go

# Final stage
FROM alpine:3.18

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /app/frontend .

EXPOSE 8080

CMD ["./frontend"]
