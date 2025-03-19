# Use a lightweight base image
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o twelve-labs-video-processor ./cmd/video_processor/main.go

# Use a minimal image for the final container
FROM alpine:3.18

RUN apk add --no-cache ffmpeg

WORKDIR /root/

COPY --from=builder /app/twelve-labs-video-processor .

EXPOSE 8080

CMD ["./twelve-labs-video-processor"]
