ARG ONOS_BUILD_VERSION=undefined

FROM onosproject/golang-build:$ONOS_BUILD_VERSION as build

RUN cd $GOPATH \
    && GO111MODULE=on go get -u \
      github.com/google/gnxi/gnoi_target@6697a080bc2d3287d9614501a3298b3dcfea06df \
      github.com/google/gnxi/gnoi_cert@6697a080bc2d3287d9614501a3298b3dcfea06df 

ENV ADAPTER_ROOT=$GOPATH/src/github.com/onosproject/sdcore-adapter
ENV CGO_ENABLED=0

RUN mkdir -p $ADAPTER_ROOT/

COPY . $ADAPTER_ROOT/

RUN cd $ADAPTER_ROOT && GO111MODULE=on go build -o /go/bin/sdcore-adapter ./cmd/sdcore-adapter

FROM alpine:3.11
RUN apk add bash openssl curl libc6-compat

ENV HOME=/home/sdcore-adapter

RUN mkdir $HOME
WORKDIR $HOME

COPY --from=build /go/bin/sdcore-adapter /usr/local/bin/
