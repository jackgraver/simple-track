FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod ./
COPY go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o server ./cmd/server

FROM alpine:3.22

WORKDIR /app

COPY --from=builder /app/server .

RUN mkdir -p /app/api /app/database/dumps

EXPOSE 8080

CMD ["./server"]
