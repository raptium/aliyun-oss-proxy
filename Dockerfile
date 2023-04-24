FROM golang:alpine AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app

FROM alpine:latest

ENV PORT 3000
EXPOSE 3000

COPY --from=builder /usr/local/bin/app /usr/local/bin/app

CMD ["app"]