module github.com/onosproject/sdcore-adapter

go 1.16

require (
	github.com/eapache/channels v1.1.0
	github.com/golang/mock v1.6.0
	github.com/golang/protobuf v1.5.2
	github.com/google/gnxi v0.0.0-20210716134716-cb5c55758a07
	github.com/gorilla/mux v1.8.0
	github.com/onosproject/aether-models/models/aether-2.0.x v0.0.0-20220327085226-9d2854b74665
	github.com/onosproject/aether-models/models/aether-2.1.x v0.0.0-20220404214232-148c0e4da437
	github.com/onosproject/aether-models/models/aether-4.x v0.0.0-20220327085226-9d2854b74665
	github.com/onosproject/analytics/pkg/kafkaClient v0.0.0-20220503194729-1cd33b3a8dc8 // indirect
	github.com/onosproject/analytics/pkg/messages v0.0.0-20220503194729-1cd33b3a8dc8
	github.com/onosproject/config-models/modelplugin/testdevice-2.0.0 v0.8.8
	github.com/onosproject/onos-lib-go v0.8.15
	github.com/openconfig/gnmi v0.0.0-20210914185457-51254b657b7d
	github.com/openconfig/goyang v0.4.0
	github.com/openconfig/ygot v0.14.0
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.26.0
	github.com/stretchr/testify v1.7.0
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd
	golang.org/x/oauth2 v0.0.0-20210819190943-2bc19b11175f
	google.golang.org/grpc v1.41.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.22.3
	k8s.io/client-go v0.22.3
)
