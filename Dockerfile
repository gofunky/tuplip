FROM gofunky/golang:1.11.0-alpine3.8 as builder

COPY . $GOPATH/src/github.com/gofunky/tuplip
WORKDIR $GOPATH/src/github.com/gofunky/tuplip/

ENV GOOS=linux
ENV GOARCH=amd64

ARG VERSION=latest

RUN dep ensure -v -vendor-only
RUN go build -v -o /go/bin/tuplip -ldflags "-X main.GitVersion=$VERSION" ./cmd/tuplip

FROM gofunky/git:2.18.1
LABEL maintainer="mat@fax.fyi"

COPY --from=builder /go/bin/tuplip /usr/local/bin/tuplip
RUN chmod +x /usr/local/bin/tuplip

ENTRYPOINT ["/usr/local/bin/tuplip"]

ARG VERSION=latest
ARG BUILD_DATE
ARG VCS_REF

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.vcs-url="https://github.com/gofunky/tuplip" \
      org.label-schema.version=$VERSION \
      org.label-schema.schema-version="1.0"
