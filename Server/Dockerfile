FROM golang:1.22-alpine

RUN mkdir -p server

WORKDIR /server

COPY . .
RUN go mod download

RUN go build -o /server

ENV HOST_NAME=localhost \
    USER_NAME=postgres \
    PASSWORD=postgres \
    DB_NAME=postgres \
    PORT=5432 \
    GOSERVER_PORT=8080 \
    MEMC_PATH=10.0.0.1:11211

EXPOSE 8080

CMD [ "./server" ]


# sudo -a -u postgres

# https://docs.docker.com/reference/dockerfile
