FROM golang:1.24 AS build-backend

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN GOOS=linux go build -o main .

FROM gcr.io/distroless/base-debian12
COPY --from=build-backend /app .
EXPOSE 8080
CMD ["./main", "serve", "--http=0.0.0.0:8080", "--dir=/cloud/storage/pb_data"]
