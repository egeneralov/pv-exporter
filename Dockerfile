FROM golang:1.14.2-alpine

RUN apk add --no-cache ca-certificates

ENV \
  GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64

WORKDIR /go/src/github.com/egeneralov/pv-exporter
ADD go.mod go.sum /go/src/github.com/egeneralov/pv-exporter/
RUN go mod download

ADD . .

RUN go build -v -installsuffix cgo -ldflags="-w -s" -o /go/bin/pv-exporter .


FROM alpine

RUN apk add --no-cache ca-certificates
USER nobody
ENV PATH='/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin'
CMD /go/bin/pv-exporter

COPY --from=0 /go/bin /go/bin

