FROM golang:1.22-alpine

RUN mkdir -p server

WORKDIR /server

COPY . .
RUN go mod download

RUN go build -o /server

EXPOSE 3000

CMD [ "./server" ]


# sudo -a -u postgres

# https://docs.docker.com/reference/dockerfile
