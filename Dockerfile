FROM golang:1.5
RUN go get github.com/tools/godep

ENV SRC_DIR=$GOPATH/src/github.com/vincentcr/tormgr/api
RUN mkdir -p /opt/app $SRC_DIR
WORKDIR $SRC_DIR
COPY . $SRC_DIR
RUN godep go build -o app .
RUN mv ./app /opt/app

EXPOSE 3456
CMD ["/opt/app/app", "-bind", ":3456"]
