FROM golang:1.19.6-alpine3.17

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -o main cmd/main.go

EXPOSE 8080

CMD [ "/app/main" ]
