FROM golang:1.25-alpine AS build

WORKDIR /app

RUN go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main cmd/api/main.go

FROM alpine:3.20.1 AS production
WORKDIR /app
COPY --from=build /go/bin/goose /usr/local/bin/goose
COPY --from=build /app/main /app/main
COPY --from=build /app/migrations /app/migrations
COPY --from=build /app/entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
EXPOSE ${PORT}
CMD ["/app/entrypoint.sh"]
