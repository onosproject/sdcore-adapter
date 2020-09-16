#!/bin/bash

hostname=${HOSTNAME:-localhost}

#CONFIG=$HOME/target_configs/typical_aether_config.json
CONFIG=$HOME/target_configs/empty_config.json

if [ $SIM_MODE == 1 ]; 
then
    if [ "$hostname" != "localhost" ]; then \
        IPADDR=`ip route get 1.2.3.4 | grep dev | awk '{print $7}'`
        $HOME/certs/generate_certs.sh $hostname > /dev/null 2>&1;
        echo "Please add '"$IPADDR" "$hostname"' to /etc/hosts and access with gNMI client at "$hostname":"$GNMI_PORT; \
    else \
        echo "gNMI target in secure mode is on $hostname:"${GNMI_PORT};
        echo "gNMI target insecure mode is on $hostname:"${GNMI_INSECURE_PORT};
    fi
    sed -i -e "s/replace-device-name/"$hostname"/g" $CONFIG && \
    sed -i -e "s/replace-motd-banner/Welcome to gNMI service on "$hostname":"$GNMI_PORT"/g" $CONFIG

    sdcore_adapter \
       -bind_address :$GNMI_INSECURE_PORT \
       -alsologtostderr \
       -notls \
       -insecure \
       -config $CONFIG \
       -output $OUTPUT &

    sdcore_adapter \
       -bind_address :$GNMI_PORT \
       -key $HOME/certs/$hostname.key \
       -cert $HOME/certs/$hostname.crt \
       -ca $HOME/certs/onfca.crt \
       -alsologtostderr \
       -config $CONFIG \
       -output $OUTPUT


elif [ $SIM_MODE == 2 ]; 
then
    if [ "$hostname" != "localhost" ]; then \
        IPADDR=`ip route get 1.2.3.4 | grep dev | awk '{print $7}'`
        echo "Please add '"$IPADDR" "$hostname"' to /etc/hosts and access with gNOI client at "$hostname":"$GNOI_PORT; \
    else \
        echo "gNOI running on $hostname:"$GNOI_PORT;
    fi    	
    gnoi_target \
      -bind_address :$GNOI_PORT \
      -alsologtostderr;
elif [ $SIM_MODE == 3 ];
then
    if [ "$hostname" != "localhost" ]; then \
        IPADDR=`ip route get 1.2.3.4 | grep dev | awk '{print $7}'`
        $HOME/certs/generate_certs.sh $hostname > /dev/null 2>&1;
        echo "Please add '"$IPADDR" "$hostname"' to /etc/hosts and access with gNMI/gNOI clients at "$hostname":"$GNMI_PORT":"$GNOI_PORT":"GNMI_INSECURE_PORT; \
    else \
        echo "gNMI target in secure mode is on $hostname:"${GNMI_PORT};
        echo "gNMI target insecure mode is on $hostname:"${GNMI_INSECURE_PORT};
        echo "gNOI running on $hostname:"$GNOI_PORT;
    fi
    sed -i -e "s/replace-device-name/"$hostname"/g" $CONFIG && \
    sed -i -e "s/replace-motd-banner/Welcome to gNMI service on "$hostname":"$GNMI_PORT"/g" $CONFIG
    sdcore_adapter \
       -bind_address :$GNMI_PORT \
       -key $HOME/certs/$hostname.key \
       -cert $HOME/certs/$hostname.crt \
       -ca $HOME/certs/onfca.crt \
       -alsologtostderr \
       -config $CONFIG \
       -output $OUTPUT &

    sdcore_adapter \
       -bind_address :$GNMI_INSECURE_PORT \
       -alsologtostderr \
       -notls \
       -insecure \
       -config $CONFIG \
       -output $OUTPUT &
    
    gnoi_target \
      -bind_address :$GNOI_PORT \
      -alsologtostderr;
fi
