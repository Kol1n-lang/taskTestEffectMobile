FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/swaggo/swag/cmd/swag@latest
RUN swag init -g ./cmd/app/main.go

RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/app/main.go

FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache postgresql-client

COPY --from=builder /app/main .
COPY --from=builder /app/migrations ./migrations
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["sh", "-c", "while ! pg_isready -h postgres -U postgres -d Subscription; do sleep 2; done && ./main"]