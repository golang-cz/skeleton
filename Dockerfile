FROM golang:1.20-alpine3.18

WORKDIR /app

COPY go.mod .

RUN go mod download

COPY . .

RUN go build -o api cmd/api/main.go

EXPOSE 8877

CMD ["./api"] 
