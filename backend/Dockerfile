FROM golang:1.22

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN GOOS=linux go build -o /app/main

WORKDIR /app/cmd

RUN GOOS=linux go build -o /cmd

EXPOSE 8080

WORKDIR /app

CMD ["/app/main"]