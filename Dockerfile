FROM golang:1.19.6-alpine3.17 AS builder
WORKDIR /app
COPY . /app
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

FROM alpine:3.17
COPY --from=builder /app/main /app/service
ENTRYPOINT [ "/app/service" ]
