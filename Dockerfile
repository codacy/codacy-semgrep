ARG TOOL_VERSION=1.80.0

# Development image used to build the codacy-semgrep wrapper
# Explicitly adding go.mod and go.sum avoids re-downloading dependencies on every build
# Go builds static binaries by default, -ldflags="-s -w" strips debug information and reduces the binary size

FROM golang:1.22-alpine3.20 as builder

WORKDIR /src

COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download

COPY cmd cmd
COPY internal internal
RUN go build -o bin/codacy-semgrep -ldflags="-s -w" ./cmd/tool

COPY .tool_version .tool_version
COPY docs /docs
RUN go run ./cmd/docgen -docFolder /docs

# Semgrep official image used to copy the semgrep binary

FROM semgrep/semgrep:$TOOL_VERSION as semgrep-cli

# Compress binaries for smaller image size

FROM alpine:3.20 as compressor

RUN apk add --no-cache upx

COPY --from=semgrep-cli /usr/local/bin/semgrep-core-proprietary /usr/local/bin/semgrep-core
# Compression seems to add flaky segmentation faults for long running processes
# RUN chmod 777 /usr/local/bin/semgrep-core && upx --lzma /usr/local/bin/semgrep-core

COPY --from=builder /src/bin/codacy-semgrep /src/bin/codacy-semgrep
RUN upx --lzma /src/bin/codacy-semgrep

# Final published image for the codacy-semgrep wrapper
# Tries to be as small as possible with only the Go static binary, the docs and the semgrep binary

FROM alpine:3.20

RUN adduser -u 2004 -D docker

COPY --from=builder --chown=docker:docker /docs /docs
COPY --from=compressor /usr/local/bin/semgrep-core /usr/bin/semgrep
COPY --from=compressor /src/bin /dist/bin
COPY --from=semgrep-cli /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD [ "/dist/bin/codacy-semgrep" ]
