FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o blockyserver .

FROM alpine:latest

RUN apk add --no-cache ffmpeg

WORKDIR /app

COPY --from=builder /app/blockyserver .

EXPOSE 8080

ENTRYPOINT ["./blockyserver"]
CMD ["-port", "8080"]
