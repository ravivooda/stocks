# syntax=docker/dockerfile:1

FROM golang:1.17

WORKDIR /app

ADD . /app
RUN go mod download

RUN go build -o /stocks .

EXPOSE 8080

CMD [ "/stocks" ]