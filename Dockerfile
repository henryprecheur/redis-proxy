FROM golang:1.10-alpine

ENV GOPATH /go
ENV WORKDIR $GOPATH/src/github.com/henryprecheur/redis-proxy
ENV PATH /go/bin:$PATH

RUN mkdir -p $WORKDIR /go/bin

WORKDIR $WORKDIR
ADD . $WORKDIR

# Setup dependencies
RUN apk add --no-cache git curl
RUN curl https://glide.sh/get | sh

RUN glide install
RUN go build
RUN ls
RUN go install

CMD ["./redis-proxy"]
