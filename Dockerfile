FROM --platform=$BUILDPLATFORM golang:1.23.3@sha256:c2d828fd49c47ed2b9192d2dbffed83052d8a21af465056d732c3de0d756f217 as builder
ARG TARGETPLATFORM
WORKDIR /usr/src/aws-ses-pop3-server
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go .
COPY e2e_test.go .
COPY pkg ./pkg
RUN go test -race -v ./...
RUN GOARCH="$(echo "${TARGETPLATFORM}" | cut -d/ -f2)" CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/aws-ses-pop3-server
FROM alpine:3.20.3@sha256:1e42bbe2508154c9126d48c2b8a75420c3544343bf86fd041fb7527e017a4b4a as runner
COPY --from=builder /usr/local/bin/aws-ses-pop3-server /usr/local/bin/aws-ses-pop3-server
RUN addgroup --gid 1767 appgroup && \
    adduser --disabled-password --gecos '' --no-create-home -G appgroup --uid 1767 appuser
USER appuser
CMD ["aws-ses-pop3-server"]
