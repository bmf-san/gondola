FROM golang:1.21.0-alpine

WORKDIR /app

COPY . .

RUN go build -o main .

CMD ["./main"]