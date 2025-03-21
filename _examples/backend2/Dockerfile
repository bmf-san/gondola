FROM golang:1.24.1-alpine

WORKDIR /app

COPY . .

RUN go build -o main .

CMD ["./main"]