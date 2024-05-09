# syntax=docker/dockerfile:1

FROM golang:1.22 AS build-stage

WORKDIR /app

COPY . .

COPY go.mod go.sum ./

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api/main.go

FROM alpine:3.19 AS build-release-stage

RUN apk --no-cache add ca-certificates bash curl

WORKDIR /app

COPY --from=build-stage /api /app/api

COPY /scripts/migrations /app/migrations

RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.1/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /app/migrate \
    && chmod +x /app/migrate

EXPOSE 8080

CMD ["/api"]
