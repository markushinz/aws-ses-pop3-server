FROM --platform=$BUILDPLATFORM golang:1.22.0@sha256:7b297d9abee021bab9046e492506b3c2da8a3722cbf301653186545ecc1e00bb as builder
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
FROM alpine:3.19.1@sha256:c5b1261d6d3e43071626931fc004f70149baeba2c8ec672bd4f27761f8e1ad6b as runner
COPY --from=builder /usr/local/bin/aws-ses-pop3-server /usr/local/bin/aws-ses-pop3-server
RUN addgroup --gid 1767 appgroup && \
    adduser --disabled-password --gecos '' --no-create-home -G appgroup --uid 1767 appuser
USER appuser
CMD ["aws-ses-pop3-server"]
