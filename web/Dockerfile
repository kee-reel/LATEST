FROM golang:latest as build
# Create dir
WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download
COPY . /app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main .

FROM scratch
COPY --from=build /app/main /bin/app
EXPOSE 1234
CMD ["/bin/app"]
