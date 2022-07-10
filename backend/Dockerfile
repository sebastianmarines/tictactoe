FROM golang:1.18-alpine as builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY *.go ./

RUN go build -o ./main

FROM alpine

ENV PORT 8080
ARG REDIS_HOST
ARG REDIS_PORT

COPY --from=builder /app/main .

ENTRYPOINT [ "./main" ]

EXPOSE 8080
