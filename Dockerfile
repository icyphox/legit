FROM golang:1.19-alpine AS builder

WORKDIR /app

COPY . ./

RUN apk add gcc musl-dev libc-dev

RUN go mod download
RUN go mod verify
RUN go build -o legit

FROM golang:1.19-alpine

WORKDIR /app

COPY static ./static
COPY templates ./templates
COPY config.yaml ./
COPY --from=builder /app/legit ./

EXPOSE 5555

CMD ["./legit"]

