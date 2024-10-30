FROM golang:1.21-alpine
LABEL authors="passwordhash"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

# На будущее
VOLUME ["/var/log"]

COPY . .

EXPOSE 2525

RUN go build -o main cmd/bot/main.go

CMD ["./main"]
