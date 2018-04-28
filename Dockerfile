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
RUN cp redis-proxy /bin
RUN go install
RUN cp ./redis-proxy.sh /
RUN chmod +x /redis-proxy.sh
RUN ls /

CMD ["/redis-proxy.sh"]
ENTRYPOINT ["/redis-proxy.sh"]
