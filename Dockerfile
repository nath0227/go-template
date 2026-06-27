FROM golang:1.25-alpine AS base
ARG USER
ARG ACCESS_TOKEN
ARG EXEC
RUN apk add --no-cache make
RUN apk add curl
RUN apk add git
RUN go env -w GOPRIVATE=git.amaze-x.com/amaze/*
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.2
WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN git config --global url."https://${USER}:${ACCESS_TOKEN}@git.amaze-x.com".insteadOf "https://git.amaze-x.com"
RUN go mod download

FROM base AS builder
COPY . ./
RUN mkdir ./build
RUN find . -name "*.json" -exec cp --parents '{}' ./build \;
RUN go build -o /app/build/app ./main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/build /app
CMD ["sh", "-c", "${EXEC}"]
