FROM golang:latest

WORKDIR /app

ENV REDIS_PASSWORD="123456"
ENV MONGO_PASSWORD="123456"

ENV GO111MODULE=on

COPY .. .

RUN cd /app/serve && go build -o /app/bin/serve
