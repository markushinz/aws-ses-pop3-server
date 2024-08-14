FROM --platform=$BUILDPLATFORM golang:1.23.0@sha256:c9718f6aa6efb2641d84faf55f9fcb7e3f048bad6e1b762382a7675a322a8d46 as builder
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
FROM alpine:3.20.2@sha256:0a4eaa0eecf5f8c050e5bba433f58c052be7587ee8af3e8b3910ef9ab5fbe9f5 as runner
COPY --from=builder /usr/local/bin/aws-ses-pop3-server /usr/local/bin/aws-ses-pop3-server
RUN addgroup --gid 1767 appgroup && \
    adduser --disabled-password --gecos '' --no-create-home -G appgroup --uid 1767 appuser
USER appuser
CMD ["aws-ses-pop3-server"]
