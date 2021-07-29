FROM onosproject/golang-build:v0.6.8 as build

ARG LOCAL_AETHER_MODELS

RUN cd $GOPATH \
    && GO111MODULE=on go get -u \
      github.com/google/gnxi/gnoi_target@6697a080bc2d3287d9614501a3298b3dcfea06df \
      github.com/google/gnxi/gnoi_cert@6697a080bc2d3287d9614501a3298b3dcfea06df 

ENV ADAPTER_ROOT=$GOPATH/src/github.com/onosproject/sdcore-adapter
ENV CGO_ENABLED=0

RUN mkdir -p $ADAPTER_ROOT/

COPY . $ADAPTER_ROOT/

# If LOCAL_AETHER_MODELS was used, then patch the go.mod file to load
# the models from the local source.
RUN if [ -n "$LOCAL_AETHER_MODELS" ] ; then \
    echo "replace github.com/onosproject/config-models/modelplugin/aether-2.1.0 => ./local-aether-models/aether-2.1.0" >> $ADAPTER_ROOT/go.mod; \
    echo "replace github.com/onosproject/config-models/modelplugin/aether-3.0.0 => ./local-aether-models/aether-3.0.0" >> $ADAPTER_ROOT/go.mod; \
    fi

RUN cat $ADAPTER_ROOT/go.mod

RUN cd $ADAPTER_ROOT && GO111MODULE=on go build -o /go/bin/sdcore-adapter ./cmd/sdcore-adapter

RUN cd $ADAPTER_ROOT && GO111MODULE=on go build -o /go/bin/sdcore-migrate ./cmd/sdcore-migrate

RUN cd $ADAPTER_ROOT && GO111MODULE=on go build -o /go/bin/sdcore-exporter ./cmd/sdcore-exporter

RUN cd $ADAPTER_ROOT && GO111MODULE=on go build -o /go/bin/sample-rocapp ./cmd/sample-rocapp

RUN cd $ADAPTER_ROOT && GO111MODULE=on go build -o /go/bin/add-imsi ./cmd/add-imsi

FROM alpine:3.11
RUN apk add bash openssl curl libc6-compat

ENV HOME=/home/sdcore-adapter

RUN mkdir $HOME
WORKDIR $HOME

COPY --from=build /go/bin/sdcore-adapter /usr/local/bin/
COPY --from=build /go/bin/sdcore-migrate /usr/local/bin/
COPY --from=build /go/bin/sdcore-exporter /usr/local/bin/
COPY --from=build /go/bin/sample-rocapp /usr/local/bin/
COPY --from=build /go/bin/add-imsi /usr/local/bin/

COPY examples/sample-rocapp.yaml /etc/
