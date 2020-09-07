FROM golang:1.11.4 as go
FROM golang:1.11.4 as i__me
FROM scratch as foo
FROM gofunky/docker:18.09.0
ARG VERSION=2.4
