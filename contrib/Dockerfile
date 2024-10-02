FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY . .
RUN go mod download
RUN go mod verify

RUN go build -o legit

FROM scratch AS build-release-stage

WORKDIR /app

COPY static ./static
COPY templates ./templates
COPY config.yaml ./
COPY --from=builder /app/legit ./

EXPOSE 5555

CMD ["./legit"]
