#!/bin/bash

if [ -z "$CONFIG" ]; then
    CONFIG_OPT=""
else
    CONFIG_OPT="-config $CONFIG"
fi

if [ "$SECURE" == "true" ]; then
    echo "gNMI target in secure mode is on $hostname:"${GNMI_PORT};
    sdcore-adapter \
        -bind_address :$GNMI_PORT \
        -key /etc/sdcore-adapter/certs/tls.key \
        -cert /etc/sdcore-adapter/certs/tls.crt \
        -ca /etc/sdcore-adapter/certs/tls.cacert \
        -alsologtostderr \
        $CONFIG_OPT \
        -output $OUTPUT
else
    echo "gNMI target insecure mode is on $hostname:"${GNMI_INSECURE_PORT};
    sdcore-adapter \
        -bind_address :$GNMI_INSECURE_PORT \
        -alsologtostderr \
        -notls \
        -insecure \
        $CONFIG_OPT \
        -output $OUTPUT
fi
