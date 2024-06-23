FROM golang:1.22.4-alpine

WORKDIR /app

COPY . .

RUN go build -o main .

CMD ["./main"]