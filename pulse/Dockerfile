FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY main.go .
COPY dashboard.html .
RUN go build -o nebula-pulse .

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/nebula-pulse .
EXPOSE 2019
CMD ["./nebula-pulse"]
