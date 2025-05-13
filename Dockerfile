FROM golang:alpine AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build cmd/main.go

FROM alpine:latest

WORKDIR /app
EXPOSE 9080

COPY config/local.yaml .

COPY --from=build /app/main /app/main

CMD ["./main"]