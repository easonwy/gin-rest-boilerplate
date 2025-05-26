# Build Stage
FROM golang:1.24.3 as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/app

# Run Stage
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/app .
COPY configs/ ./configs/

CMD ["./app"]