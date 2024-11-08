FROM golang:1.21-alpine AS base

WORKDIR /app

# Build
FROM base AS build

COPY --link go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main cmd/bot/main.go

# Run
FROM base

COPY .env /app/.env
COPY proxies.json /app/proxies.json

COPY --from=build /app/main /app/main
COPY --from=build /app/assets /app/assets

VOLUME /app/logs

CMD ["./main"]
