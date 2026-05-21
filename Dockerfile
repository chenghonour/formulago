FROM golang:1.26-alpine AS builder

LABEL maintainer="coko@duck.com"

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o formulago .

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

ENV WORKDIR /var/www/formulago
ENV IS_PROD false

WORKDIR $WORKDIR
COPY --from=builder /src/formulago .

EXPOSE 8191

CMD ./formulago
