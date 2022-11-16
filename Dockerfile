# SPDX-FileCopyrightText: 2022-present Intel Corporation
# SPDX-FileCopyrightText: 2019-present Open Networking Foundation <info@opennetworking.org>
#
# SPDX-License-Identifier: Apache-2.0

FROM onosproject/golang-build:v1.0 as build

ARG LOCAL_AETHER_MODELS
ARG org_label_schema_version=unknown
ARG org_label_schema_vcs_url=unknown
ARG org_label_schema_vcs_ref=unknown
ARG org_label_schema_build_date=unknown
ARG org_opencord_vcs_commit_date=unknown
ARG org_opencord_vcs_dirty=unknown

ENV ADAPTER_ROOT=$GOPATH/src/github.com/onosproject/sdcore-adapter
ENV CGO_ENABLED=0

RUN mkdir -p $ADAPTER_ROOT/

COPY . $ADAPTER_ROOT/

# If LOCAL_AETHER_MODELS was used, then patch the go.mod file to load
# the models from the local source.
RUN if [ -n "$LOCAL_AETHER_MODELS" ] ; then \
    echo "replace github.com/onosproject/aether-models/models/aether-4.x => ./local-aether-models/aether-4.x" >> $ADAPTER_ROOT/go.mod; \
    echo "replace github.com/onosproject/aether-models/models/aether-2.0.x => ./local-aether-models/aether-2.0.x" >> $ADAPTER_ROOT/go.mod; \
    fi

RUN cat $ADAPTER_ROOT/go.mod

RUN cd $ADAPTER_ROOT && GO111MODULE=on go build -o /go/bin/sdcore-adapter \
        -ldflags \
        "-X github.com/onosproject/sdcore-adapter/internal/pkg/version.Version=$org_label_schema_version \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.GitCommit=$org_label_schema_vcs_ref  \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.GitDirty=$org_opencord_vcs_dirty \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.GoVersion=$(go version 2>&1 | sed -E  's/.*go([0-9]+\.[0-9]+\.[0-9]+).*/\1/g') \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.Os=$(go env GOHOSTOS) \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.Arch=$(go env GOHOSTARCH) \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.BuildTime=$org_label_schema_build_date" \
         ./cmd/sdcore-adapter

RUN cd $ADAPTER_ROOT && GO111MODULE=on go build -o /go/bin/sdcore-migrate \
        -ldflags \
        "-X github.com/onosproject/sdcore-adapter/internal/pkg/version.Version=$org_label_schema_version \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.GitCommit=$org_label_schema_vcs_ref  \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.GitDirty=$org_opencord_vcs_dirty \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.GoVersion=$(go version 2>&1 | sed -E  's/.*go([0-9]+\.[0-9]+\.[0-9]+).*/\1/g') \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.Os=$(go env GOHOSTOS) \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.Arch=$(go env GOHOSTARCH) \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.BuildTime=$org_label_schema_build_date" \
         ./cmd/sdcore-migrate

RUN cd $ADAPTER_ROOT && GO111MODULE=on go build -o /go/bin/sdcore-kafka \
        -ldflags \
        "-X github.com/onosproject/sdcore-adapter/internal/pkg/version.Version=$org_label_schema_version \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.GitCommit=$org_label_schema_vcs_ref  \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.GitDirty=$org_opencord_vcs_dirty \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.GoVersion=$(go version 2>&1 | sed -E  's/.*go([0-9]+\.[0-9]+\.[0-9]+).*/\1/g') \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.Os=$(go env GOHOSTOS) \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.Arch=$(go env GOHOSTARCH) \
         -X github.com/onosproject/sdcore-adapter/internal/pkg/version.BuildTime=$org_label_schema_build_date" \
         ./cmd/sdcore-kafka

FROM alpine:3.11
RUN apk add bash openssl curl libc6-compat

ENV HOME=/home/sdcore-adapter

RUN mkdir $HOME
WORKDIR $HOME

COPY --from=build /go/bin/sdcore-adapter /usr/local/bin/
RUN cd /usr/local/bin && ln -s sdcore-adapter roc-adapter && cd $HOME
COPY --from=build /go/bin/sdcore-migrate /usr/local/bin/
COPY --from=build /go/bin/sdcore-kafka /usr/local/bin/

COPY examples/sample-rocapp.yaml /etc/
