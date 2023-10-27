FROM golang:1.21.0-alpine3.18 as builder

WORKDIR /src
RUN apk add python3 py3-pip
RUN python3 -m pip install semgrep

ADD . .
RUN go build -o bin/codacy-semgrep

FROM alpine:3.18

ENV PATH="/go/bin:${PATH}"
COPY --from=builder /go /go
COPY --from=builder /src/bin /dist/bin
# COPY docs/ /docs/

RUN adduser -u 2004 -D docker
# RUN chown -R docker:docker /docs

CMD [ "/dist/bin/codacy-semgrep" ]
