FROM golang:1.25-bookworm

RUN apt-get update && apt-get install -y librdkafka-dev && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o worker .

CMD ["./worker"]
