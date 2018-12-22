FROM  golang:lates

RUN apt update -y -qq && \
    apt install mysql-cli

ENV GO111MODULE=on

RUN mkdir -p /go/src/grpc-tutorial
ADD . /go/src/grpc-tutorial

RUN go get -u
RUN go build
EXPOSE 3000