FROM golang:1.23-alpine AS build
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o api ./cmd/api

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=build /app/api /api
ENTRYPOINT ["/api"]
