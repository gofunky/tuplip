FROM gofunky/golang:1.11.0 as i__builder

COPY . $GOPATH/src/github.com/gofunky/tuplip
WORKDIR $GOPATH/src/github.com/gofunky/tuplip/

ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on

ARG VERSION=latest

RUN go get -v github.com/ahmetb/govvv
RUN govvv build -v -o /go/bin/tuplip ./cmd/tuplip
RUN go test -v ./...

FROM docker:18.09.1-git
LABEL maintainer="mat@fax.fyi"

COPY --from=i__builder /go/bin/tuplip /usr/local/bin/tuplip
RUN chmod +x /usr/local/bin/tuplip

ENTRYPOINT ["/usr/local/bin/tuplip"]

ARG VERSION=latest
ARG BUILD_DATE
ARG VCS_REF

LABEL org.label-schema.build-date=$BUILD_DATE \
      org.label-schema.vcs-url="https://github.com/gofunky/tuplip" \
      org.label-schema.version=$VERSION \
      org.label-schema.schema-version="1.0"
