# syntax=docker/dockerfile:1.7

FROM golang:1.26.3-bookworm AS dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
EXPOSE 8080
CMD ["go", "run", "."]
