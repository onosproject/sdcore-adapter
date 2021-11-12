// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

// Ideally we would get these from the yang defaults
const (
	// DefaultAdminStatus is the default for the AdminStatus Field
	DefaultAdminStatus = "ENABLE"

	// DefaultMTU is the default for the MTU field
	DefaultMTU = 1492

	// DefaultProtocol is the default for the Protocol field
	DefaultProtocol = "TCP"

	// DefaultBitrateUnit is the unit we use for bitrates
	DefaultBitrateUnit = "bps"

	// DefaultUplinkBurst is the default Uplink Burst Size for Slice MBR
	DefaultUplinkBurst = 625000

	// DefaultDownlinkBurst is the default Downlink Burst Size for Slice MBR
	DefaultDownlinkBurst = 625000
)

// The following structures define the JSON schema used by the SD-Core.

type trafficClass struct {
	Name string `json:"name"`
	QCI  uint8  `json:"qci"`
	ARP  uint8  `json:"arp"`
	PDB  uint16 `json:"pdb"`
	PELR uint8  `json:"pelr"`
}

type ipdQos struct {
	Uplink       uint64        `json:"dnn-mbr-uplink"`
	Downlink     uint64        `json:"dnn-mbr-downlink"`
	Unit         *string       `json:"bitrate-unit"`
	TrafficClass *trafficClass `json:"traffic-class,omitempty"`
}

type ipDomain struct {
	Dnn          string  `json:"dnn"`
	Pool         string  `json:"ue-ip-pool"`
	DNSPrimary   string  `json:"dns-primary"`
	DNSSecondary string  `json:"dns-secondary,omitempty"`
	Mtu          uint16  `json:"mtu"`
	Qos          *ipdQos `json:"ue-dnn-qos,omitempty"`
}

type deviceGroup struct {
	Imsis        []string `json:"imsis"`
	IPDomainName string   `json:"ip-domain-name"`
	SiteInfo     string   `json:"site-info"`
	IPDomain     ipDomain `json:"ip-domain-expanded"`
}

type sliceIDStruct struct {
	Sst string `json:"sst"`
	Sd  string `json:"sd"`
}

type gNodeB struct {
	Name string `json:"name"`
	Tac  uint32 `json:"tac"`
}

type plmn struct {
	Mcc string `json:"mcc"`
	Mnc string `json:"mnc"`
}

type upf struct {
	Name string `json:"upf-name"`
	Port uint16 `json:"upf-port"`
}

type siteInfo struct {
	SiteName string   `json:"site-name"`
	Plmn     plmn     `json:"plmn"`
	GNodeBs  []gNodeB `json:"gNodeBs"`
	Upf      upf      `json:"upf"`
}

type appFilterRule struct {
	Name          string        `json:"rule-name"`
	Priority      uint8         `json:"priority"`
	Action        string        `json:"action"`
	Endpoint      string        `json:"endpoint"`
	DestPortStart *uint16       `json:"dest-port-start,omitempty"`
	DestPortEnd   *uint16       `json:"dest-port-end,omitempty"`
	Protocol      *uint8        `json:"protocol,omitempty"`
	Uplink        uint64        `json:"app-mbr-uplink,omitempty"`
	Downlink      uint64        `json:"app-mbr-downlink,omitempty"`
	Unit          *string       `json:"bitrate-unit,omitempty"`
	TrafficClass  *trafficClass `json:"traffic-class,omitempty"`
}

type slice struct {
	ID                        sliceIDStruct   `json:"slice-id"`
	DeviceGroup               []string        `json:"site-device-group,omitempty"`
	SiteInfo                  siteInfo        `json:"site-info"`
	ApplicationFilteringRules []appFilterRule `json:"application-filtering-rules"`
}
