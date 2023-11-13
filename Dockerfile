FROM --platform=$BUILDPLATFORM golang:1.21.4@sha256:81cd210ae58a6529d832af2892db822b30d84f817a671b8e1c15cff0b271a3db as builder
ARG TARGETPLATFORM
ENV GO111MODULE=on
WORKDIR /usr/src/aws-ses-pop3-server
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY main.go .
COPY e2e_test.go .
COPY pkg ./pkg
RUN go test -race -v ./...
RUN GOARCH="$(echo "${TARGETPLATFORM}" | cut -d/ -f2)" CGO_ENABLED=0 GOOS=linux go build -o /usr/local/bin/aws-ses-pop3-server
FROM alpine:3.18.4@sha256:eece025e432126ce23f223450a0326fbebde39cdf496a85d8c016293fc851978 as runner
COPY --from=builder /usr/local/bin/aws-ses-pop3-server /usr/local/bin/aws-ses-pop3-server
RUN addgroup --gid 1767 appgroup && \
    adduser --disabled-password --gecos '' --no-create-home -G appgroup --uid 1767 appuser
USER appuser
CMD ["aws-ses-pop3-server"]
