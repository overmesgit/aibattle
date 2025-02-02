FROM golang:1.23 AS build-backend

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest AS production
COPY --from=build-backend /app .
EXPOSE 8080
CMD ["./main", "serve", "--http=0.0.0.0:8080", "--dir=/cloud/storage/pb_data"]