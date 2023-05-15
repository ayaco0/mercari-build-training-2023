FROM golang:1.20-alpine

WORKDIR /app
COPY db db
COPY go go
EXPOSE 9000

RUN addgroup -S mercari && adduser -S trainee -G mercari
RUN chown -R trainee:mercari /app/db && chown -R trainee:mercari /app/go/images
RUN apk add --no-cache git sqlite-dev gcc musl-dev

WORKDIR /app/go
RUN go mod download

USER trainee
WORKDIR /app/go/app
CMD ["go", "run", "main.go"]