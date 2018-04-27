FROM golang:1.10

ADD . .

RUN \
  apk add --no-cache bash git openssh curl && \
  curl https://glide.sh/get | sh && \
  glide install

CMD ["ls"]
