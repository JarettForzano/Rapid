FROM golang:latest as build
COPY . .
WORKDIR /src
RUN go run main.go