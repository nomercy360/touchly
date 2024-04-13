# syntax=docker/dockerfile:1

# Build the application from source
FROM golang:1.22 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /api ./cmd/api

# Deploy the application binary into a lean image
FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build-stage /api /api

EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/api"]