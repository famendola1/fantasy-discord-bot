# syntax=docker/dockerfile:1

FROM golang:1.19-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY bot/ ./bot/
COPY providers/ ./providers/
COPY conf.json ./

RUN go build -o fantasy_bot bot/main.go

CMD [ "./fantasy_bot", "--cfg=conf.json" ]
