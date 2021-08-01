FROM golang:alpine

WORKDIR /app

COPY . .

RUN go mod download
RUN go build -o main ./cmd
RUN cat .env
CMD ["./main"]