FROM golang:latest as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN apk update
RUN apk upgrade
RUN apk add bash
WORKDIR /root/
COPY --from=builder /app .

EXPOSE 1234

CMD ["./main"]
