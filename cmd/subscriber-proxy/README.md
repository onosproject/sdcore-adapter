<!--
SPDX-FileCopyrightText: 2019-present Open Networking Foundation <info@opennetworking.org>

SPDX-License-Identifier: Apache-2.0
-->

# Subscribe Proxy 

This component acts as proxy between the sdcore simapp and config pods config4g in case of 4G or 
webui in case of 5g in case of ROC component is there. So, that ROC is aware of the imsi allocation 
and update the device group information

- SimApp --> config4g/webui (wihout ROC)
- SimApp --> subscriber-proxy --> config4g/webui (in case of ROC)

Run with:

```bash
go run ./cmd/subscriber-proxy/subscriber-proxy --bind_port :5001 \
--onos_config_url  onos-config:5150 \ 
-hostCheckDisabled -client_key=tls.key -client_crt=tls.crt -ca_crt=tls.cacert 
```

Note : 
1. Currently there is no option to run outside of the k8s cluster as it's using k8s apis
2. For Security (Auth enabled) set environment variable with relevent auth provider url. 

```
OIDC_SERVER_URL=http://k3u-keycloak:5557/auth/realms/master
``` 

## Deploying subscriber proxy using aether-in-a-box

####4G 

```
make roc-4g-model
make omec
```
####5G

```
make roc-5g-model
make 5gc
```
#### Security enabled

Update roc-values-v4.yaml OIDC issuer url as below

```
onos-config:
  openidc:
    issuer: http://k3u-keycloak:80/auth/realms/master
  models:
    plproxy:
      v1:
        enabled: true
aether-roc-gui-v4:
  ingress:
    enabled: false
  openidc:
    issuer: http://k3u-keycloak:5557/auth/realms/master

aether-roc-api:
  openidc:
    issuer: http://k3u-keycloak:80/auth/realms/master
subscriber-proxy:
  config:
    openidc:
      issuer: http://k3u-keycloak:80/auth/realms/master   
```

###Deploying subscriber proxy using aether-roc-umbrella

It will sdcore-dummy-test instead of config4/webui component 
so, that we can test it without the sdcore component

#### Testing using curl command

```
curl --location --request POST 'http://subscriber-proxy.aether-roc.svc.cluster.local:5000/api/subscriber/imsi-208014567891201' \
--header 'Content-Type: text/plain' \
--data-raw '{
    "plmnID": "20893",
    "ueId": "imsi-208014567891201",
    "OPc": "8e27b6af0e692e750f32667a3b14605d",
    "key": "8baf473f2f8fd09487cccbd7097c6862",
    "sequenceNumber":"16f3b3f70fc2",
    "DNN": "internet "	
}'
```


