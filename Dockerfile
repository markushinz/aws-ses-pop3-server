FROM golang:1.13 as builder
WORKDIR /usr/src/aws-ses-pop3-server
COPY go.mod .
COPY go.sum .
COPY main.go .
COPY pkg ./pkg
RUN go mod download
RUN go test -race -v ./...
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o /usr/local/bin/aws-ses-pop3-server
FROM alpine:3.11 as runner
COPY --from=builder /usr/local/bin/aws-ses-pop3-server /usr/local/bin/aws-ses-pop3-server
CMD ["aws-ses-pop3-server"]