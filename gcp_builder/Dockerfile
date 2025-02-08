FROM golang:1.23

WORKDIR /app

COPY main.go .

RUN go mod init builder

RUN go build -o compiler .

EXPOSE 8080

CMD ["./compiler"]
