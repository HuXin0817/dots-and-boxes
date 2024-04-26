FROM golang:latest

WORKDIR /app

ENV REDIS_PASSWORD="123456"

ENV GO111MODULE=on

COPY .. .

RUN cd /app/engine && go build -o /app/bin/engine
