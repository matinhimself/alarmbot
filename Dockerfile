FROM golang:alpine


COPY /go.mod app/
COPY /go.sum app/

WORKDIR app/
RUN go mod download
COPY . .

RUN go build -o main ./cmd

CMD ["./main"]