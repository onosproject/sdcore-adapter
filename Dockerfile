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

RUN cd $ADAPTER_ROOT && GO111MODULE=on go build -o /go/bin/sdcore_adapter ./cmd/sdcore_adapter

FROM alpine:3.11
RUN apk add bash openssl curl libc6-compat
ENV GNMI_PORT=10161
ENV GNMI_INSECURE_PORT=11161
ENV GNOI_PORT=50001
ENV SIM_MODE=1
ENV HOME=/home/sdcore-adapter
ENV OUTPUT=/home/sdcore-adapter/output.json
RUN mkdir $HOME
WORKDIR $HOME

COPY --from=build /go/bin/sdcore_adapter /usr/local/bin/

COPY configs/target_configs target_configs
COPY tools/scripts scripts
COPY pkg/certs certs

RUN chmod +x ./scripts/run_targets.sh
CMD ["./scripts/run_targets.sh"]
