FROM golang:1.17.3 as builder
ENV GO111MODULE=on
WORKDIR /usr/src/aws-ses-pop3-server
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go .
COPY e2e_test.go .
COPY pkg ./pkg
RUN go test -race -v ./...
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o /usr/local/bin/aws-ses-pop3-server
FROM alpine:3.14.2 as runner
COPY --from=builder /usr/local/bin/aws-ses-pop3-server /usr/local/bin/aws-ses-pop3-server
RUN addgroup --gid 1767 appgroup && \
    adduser --disabled-password --gecos '' --no-create-home -G appgroup --uid 1767 appuser
USER appuser
CMD ["aws-ses-pop3-server"]
