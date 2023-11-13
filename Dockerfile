ARG TOOL_VERSION

# Development image used to build the codacy-semgrep wrapper
# Explicitly adding go.mod and go.sum avoids re-downloading dependencies on every build
# Go builds static binaries by default, -ldflags="-s -w" strips debug information and reduces the binary size

FROM golang:1.21.0-alpine3.18 as builder

WORKDIR /src

ADD go.mod go.mod
ADD go.sum go.sum
RUN go mod download

ADD cmd cmd
ADD internal internal

RUN go build -o bin/codacy-semgrep -ldflags="-s -w" ./cmd/tool

ADD .tool_version .tool_version
COPY docs/ /docs/
RUN go run ./cmd/docgen -docFolder /docs

# Semgrep official image used to copy the semgrep binary

FROM returntocorp/semgrep:$TOOL_VERSION as semgrep-cli

# Final published image for the codacy-semgrep wrapper
# Tries to be as small as possible with only the Go static binary, the docs and the semgrep binary

FROM busybox

COPY --from=semgrep-cli /usr/local/bin/semgrep-core /usr/bin/semgrep
COPY --from=semgrep-cli /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /src/bin /dist/bin
COPY --from=builder /docs/ /docs/
COPY auto.yaml /


RUN adduser -u 2004 -D docker
RUN chown -R docker:docker /docs

CMD [ "/dist/bin/codacy-semgrep" ]
