# Use a lightweight base image
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

ENV GOOS=linux
ENV GOARCH=arm64
ENV CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o twelve-labs-api ./cmd/api/main.go

# Use a minimal image for the final container
FROM alpine:3.18
WORKDIR /root/

COPY --from=builder /app/twelve-labs-api .

EXPOSE 8080

CMD ["./twelve-labs-api"]
