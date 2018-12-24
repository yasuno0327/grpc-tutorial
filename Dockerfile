FROM  golang:lates

RUN apt update -y -qq
    # 今回は使わないapt install mysql-cli

# Install protoc command
ENV PROTOC_ZIP=protoc-3.3.0-linux-x86_64.zip
RUN curl -OL https://github.com/google/protobuf/releases/download/v3.3.0/$PROTOC_ZIP && \
    sudo unzip -o $PROTOC_ZIP -d /usr/local bin/protoc
RUN rm -f $PROTOC_ZIP

ENV GO111MODULE=on

RUN mkdir -p /go/src/grpc-tutorial
ADD . /go/src/grpc-tutorial

# Compile proto file
# RUN protoc -I proto/ --go_out=plugins=grpc:proto proto/route_guide.proto

RUN go get -u
RUN go build
EXPOSE 3000