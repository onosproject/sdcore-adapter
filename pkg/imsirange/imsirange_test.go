// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0
package imsirange

import (
	"context"
	"github.com/golang/mock/gomock"
	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	"github.com/onosproject/sdcore-adapter/pkg/test/mocks"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
	"testing"
)

var irange = ImsiRange{
	AetherConfigAddress: "onos-config.micro-onos.svc.cluster.local:5150",
	AetherConfigTarget:  "connectivity-service-v3",
}

func TestImsiRange_GetDevice(t *testing.T) {

	ctrl := gomock.NewController(t)
	gnmiClient := mocks.NewMockGnmiInterface(ctrl)
	gnmiClient.EXPECT().GetPath(gomock.Any(), "", "connectivity-service-v3",
		"onos-config.micro-onos.svc.cluster.local:5150").
		DoAndReturn(func(ctx context.Context, path string, target string, addr string) (*gpb.TypedValue, error) {
			return &gpb.TypedValue{
				Value: &gpb.TypedValue_StringVal{StringVal: `{
  "ap-list": {
    "ap-list": [
      {
        "access-points": [
          {
            "address": "ap2.chicago.acme.com",
            "enable": true,
            "tac": 8002
          }
        ],
        "description": "Chicago APs",
        "display-name": "Chicago",
        "enterprise": "acme",
        "id": "acme-chicago-aps"
      },
      {
        "access-points": [
          {
            "address": "ap2.newyork.starbucks.com",
            "enable": true,
            "tac": 8002
          }
        ],
        "description": "New York APs",
        "display-name": "New York",
        "enterprise": "starbucks",
        "id": "starbucks-newyork-aps"
      },
      {
        "access-points": [
          {
            "address": "ap1.seattle.starbucks.com",
            "enable": true,
            "tac": 654
          },
          {
            "address": "ap2.seattle.starbucks.com",
            "enable": true,
            "tac": 87475
          }
        ],
        "description": "Seattle APs",
        "display-name": "Seattle",
        "enterprise": "starbucks",
        "id": "starbucks-seattle-aps"
      }
    ]
  },
  "application": {
    "application": [
      {
        "description": "Data Acquisition",
        "display-name": "DA",
        "endpoint": [
          {
            "address": "da.acme.com",
            "name": "da",
            "port-end": 7588,
            "port-start": 7585,
            "protocol": "TCP"
          }
        ],
        "enterprise": "acme",
        "id": "acme-dataacquisition"
      },
      {
        "description": "Fidelio POS",
        "display-name": "Fidelio",
        "endpoint": [
          {
            "address": "fidelio.starbucks.com",
            "name": "fidelio",
            "port-end": 7588,
            "port-start": 7585,
            "protocol": "TCP"
          }
        ],
        "enterprise": "starbucks",
        "id": "starbucks-fidelio"
      },
      {
        "description": "Network Video Recorder",
        "display-name": "NVR",
        "endpoint": [
          {
            "address": "nvr.starbucks.com",
            "name": "rtsp",
            "port-end": 3330,
            "port-start": 3316,
            "protocol": "UDP"
          }
        ],
        "enterprise": "starbucks",
        "id": "starbucks-nvr"
      }
    ]
  },
  "connectivity-service": {
    "connectivity-service": [
      {
        "description": "ROC 4G Test Connectivity Service",
        "display-name": "4G Test",
        "hss-endpoint": "http://aether-roc-umbrella-sdcore-test-dummy/v1/config/imsis",
        "id": "cs4gtest",
        "pcrf-endpoint": "http://aether-roc-umbrella-sdcore-test-dummy/v1/config/policies",
        "spgwc-endpoint": "http://aether-roc-umbrella-sdcore-test-dummy/v1/config"
      },
      {
        "core-5g-endpoint": "http://aether-roc-umbrella-sdcore-test-dummy/v1/config/5g",
        "description": "5G Test",
        "display-name": "ROC 5G Test Connectivity Service",
        "id": "cs5gtest"
      }
    ]
  },
  "device-group": {
    "device-group": [
      {
        "display-name": "ACME Default",
        "id": "acme-chicago-default",
        "ip-domain": "acme-chicago",
        "site": "acme-chicago"
      },
      {
        "display-name": "ACME Robots",
        "id": "acme-chicago-robots",
        "imsis": [
          {
            "imsi-range-to": "3",
            "name": "production"
          },
          {
            "imsi-range-from": "10",
            "imsi-range-to": "12",
            "name": "warehouse"
          }
        ],
        "ip-domain": "acme-chicago",
        "site": "acme-chicago"
      },
      {
        "display-name": "Global Default Device Group",
        "id": "defaultent-defaultsite-default",
        "imsis": [
          {
            "imsi-range-from": "111222333000010",
            "imsi-range-to": "111222333000014",
            "name": "auto-111222333000010"
          },
          {
            "imsi-range-from": "111222333000015",
            "name": "auto-111222333000015"
          }
        ],
        "ip-domain": "defaultent-defaultip",
        "site": "defaultent-defaultsite"
      },
      {
        "display-name": "New York Cameras",
        "id": "starbucks-newyork-cameras",
        "imsis": [
          {
            "imsi-range-from": "40",
            "imsi-range-to": "41",
            "name": "front"
          },
          {
            "imsi-range-from": "50",
            "imsi-range-to": "55",
            "name": "store"
          }
        ],
        "ip-domain": "starbucks-newyork",
        "site": "starbucks-newyork"
      },
      {
        "display-name": "New York Default",
        "id": "starbucks-newyork-default",
        "imsis": [
          {
            "imsi-range-from": "94",
            "name": "auto-21032002000094"
          },
          {
            "imsi-range-from": "8",
            "imsi-range-to": "11",
            "name": "auto-21322-8"
          },
          {
            "imsi-range-from": "91",
            "imsi-range-to": "93",
            "name": "auto-21322-91"
          }
        ],
        "ip-domain": "starbucks-newyork",
        "site": "starbucks-newyork"
      },
      {
        "display-name": "New York POS",
        "id": "starbucks-newyork-pos",
        "imsis": [
          {
            "imsi-range-from": "70",
            "imsi-range-to": "73",
            "name": "store"
          },
          {
            "imsi-range-from": "60",
            "imsi-range-to": "61",
            "name": "tills"
          }
        ],
        "ip-domain": "starbucks-newyork",
        "site": "starbucks-newyork"
      },
      {
        "display-name": "Seattle Cameras",
        "id": "starbucks-seattle-cameras",
        "imsis": [
          {
            "imsi-range-to": "3",
            "name": "counters"
          },
          {
            "imsi-range-from": "10",
            "imsi-range-to": "14",
            "name": "store"
          }
        ],
        "ip-domain": "starbucks-seattle",
        "site": "starbucks-seattle"
      },
      {
        "display-name": "Seattle Default",
        "id": "starbucks-seattle-default",
        "ip-domain": "starbucks-seattle",
        "site": "starbucks-seattle"
      },
      {
        "display-name": "Seattle POS",
        "id": "starbucks-seattle-pos",
        "imsis": [
          {
            "imsi-range-from": "30",
            "imsi-range-to": "34",
            "name": "store"
          },
          {
            "imsi-range-from": "20",
            "imsi-range-to": "22",
            "name": "tills"
          }
        ],
        "ip-domain": "starbucks-seattle",
        "site": "starbucks-seattle"
      }
    ]
  },
  "enterprise": {
    "enterprise": [
      {
        "connectivity-service": [
          {
            "connectivity-service": "cs5gtest",
            "enabled": true
          }
        ],
        "description": "ACME Corporation",
        "display-name": "ACME Corp",
        "id": "acme"
      },
      {
        "description": "This Enterprise holds discovered IMSIs that cannot be associated elsewhere.",
        "display-name": "Default Enterprise",
        "id": "defaultent"
      },
      {
        "connectivity-service": [
          {
            "connectivity-service": "cs4gtest",
            "enabled": false
          },
          {
            "connectivity-service": "cs5gtest",
            "enabled": true
          }
        ],
        "description": "Starbucks Corporation",
        "display-name": "Starbucks Inc.",
        "id": "starbucks"
      }
    ]
  },
  "ip-domain": {
    "ip-domain": [
      {
        "admin-status": "DISABLE",
        "description": "Chicago IP Domain",
        "display-name": "Chicago",
        "dns-primary": "8.8.8.4",
        "dns-secondary": "8.8.8.4",
        "enterprise": "acme",
        "id": "acme-chicago",
        "mtu": 12690,
        "subnet": "163.25.44.0/31"
      },
      {
        "admin-status": "ENABLE",
        "description": "Global Default IP Domain",
        "display-name": "Global Default IP Domain",
        "dns-primary": "8.8.8.1",
        "dns-secondary": "8.8.8.2",
        "enterprise": "defaultent",
        "id": "defaultent-defaultip",
        "mtu": 57600,
        "subnet": "192.168.0.0/24"
      },
      {
        "admin-status": "ENABLE",
        "description": "New York IP Domain",
        "display-name": "New York",
        "dns-primary": "8.8.8.1",
        "dns-secondary": "8.8.8.2",
        "enterprise": "starbucks",
        "id": "starbucks-newyork",
        "mtu": 57600,
        "subnet": "254.186.117.251/31"
      },
      {
        "admin-status": "ENABLE",
        "description": "Seattle IP Domain",
        "display-name": "Seattle",
        "dns-primary": "8.8.8.3",
        "dns-secondary": "8.8.8.3",
        "enterprise": "starbucks",
        "id": "starbucks-seattle",
        "mtu": 12690,
        "subnet": "196.5.91.0/31"
      }
    ]
  },
  "site": {
    "site": [
      {
        "description": "ACME HQ",
        "display-name": "Chicago",
        "enterprise": "acme",
        "id": "acme-chicago",
        "imsi-definition": {
          "enterprise": 1,
          "format": "CCCNNNEEESSSSSS",
          "mcc": 123,
          "mnc": 456
        }
      },
      {
        "description": "Global Default Site",
        "display-name": "Global Default Site",
        "enterprise": "defaultent",
        "id": "defaultent-defaultsite",
        "imsi-definition": {
          "format": "SSSSSSSSSSSSSSS"
        }
      },
      {
        "description": "Starbucks New York",
        "display-name": "New York",
        "enterprise": "starbucks",
        "id": "starbucks-newyork",
        "imsi-definition": {
          "enterprise": 2,
          "format": "CCCNNNEEESSSSSS",
          "mcc": 21,
          "mnc": 32
        }
      },
      {
        "description": "Starbucks Corp HQ",
        "display-name": "Seattle",
        "enterprise": "starbucks",
        "id": "starbucks-seattle",
        "imsi-definition": {
          "enterprise": 2,
          "format": "CCCNNNEEESSSSSS",
          "mcc": 265,
          "mnc": 122
        }
      }
    ]
  },
  "template": {
    "template": [
      {
        "description": "VCS Template 1",
        "display-name": "Template 1",
        "downlink": 5,
        "id": "template-1",
        "sd": 10886763,
        "sst": 158,
        "traffic-class": "class-1",
        "uplink": 10
      },
      {
        "description": "VCS Template 2",
        "display-name": "Template 2",
        "downlink": 5,
        "id": "template-2",
        "sd": 16619900,
        "sst": 157,
        "traffic-class": "class-2",
        "uplink": 10
      }
    ]
  },
  "traffic-class": {
    "traffic-class": [
      {
        "description": "High Priority TC",
        "display-name": "Class 1",
        "id": "class-1",
        "pdb": 577,
        "pelr": 3,
        "qci": 10
      },
      {
        "description": "Medium Priority TC",
        "display-name": "Class 2",
        "id": "class-2",
        "pdb": 831,
        "pelr": 4,
        "qci": 20
      },
      {
        "description": "Low Priority TC",
        "display-name": "Class 3",
        "id": "class-3",
        "pdb": 833,
        "pelr": 4,
        "qci": 30
      }
    ]
  },
  "upf": {
    "upf": [
      {
        "address": "chicago.robots-upf.acme.com",
        "description": "Chicago Robots UPF",
        "display-name": "Chicago Robots",
        "enterprise": "acme",
        "id": "acme-chicago-robots",
        "port": 6161
      },
      {
        "address": "newyork.cameras-upf.starbucks.com",
        "description": "New York Cameras UPF",
        "display-name": "New York Cameras",
        "enterprise": "starbucks",
        "id": "starbucks-newyork-cameras",
        "port": 6161
      },
      {
        "address": "newyork.pos-upf.starbucks.com",
        "description": "NewYork POS UPF",
        "display-name": "NewYork POS",
        "enterprise": "starbucks",
        "id": "starbucks-newyork-pos",
        "port": 6161
      },
      {
        "address": "seattle.cameras-upf.starbucks.com",
        "description": "Seattle Cameras UPF",
        "display-name": "Seattle Cameras",
        "enterprise": "starbucks",
        "id": "starbucks-seattle-cameras",
        "port": 9229
      }
    ]
  },
  "vcs": {
    "vcs": [
      {
        "ap": "acme-chicago-aps",
        "application": [
          {
            "allow": false,
            "application": "acme-dataacquisition"
          }
        ],
        "description": "Chicago Robots",
        "device-group": [
          {
            "device-group": "acme-chicago-robots",
            "enable": true
          }
        ],
        "display-name": "Chicago Robots VCS",
        "downlink": 10,
        "enterprise": "acme",
        "id": "acme-chicago-robots",
        "sd": 2973238,
        "sst": 79,
        "template": "template-2",
        "traffic-class": "class-2",
        "upf": "acme-chicago-robots",
        "uplink": 5
      },
      {
        "ap": "starbucks-newyork-aps",
        "application": [
          {
            "allow": true,
            "application": "starbucks-nvr"
          }
        ],
        "description": "New York Cameras",
        "device-group": [
          {
            "device-group": "starbucks-newyork-cameras",
            "enable": true
          }
        ],
        "display-name": "NY Cams",
        "downlink": 10,
        "enterprise": "starbucks",
        "id": "starbucks-newyork-cameras",
        "sd": 8284729,
        "sst": 127,
        "template": "template-1",
        "traffic-class": "class-1",
        "upf": "starbucks-newyork-cameras",
        "uplink": 10
      },
      {
        "ap": "starbucks-seattle-aps",
        "application": [
          {
            "allow": false,
            "application": "starbucks-nvr"
          }
        ],
        "description": "Seattle Cameras",
        "device-group": [
          {
            "device-group": "starbucks-seattle-cameras",
            "enable": true
          }
        ],
        "display-name": "Seattle Cams",
        "downlink": 10,
        "enterprise": "starbucks",
        "id": "starbucks-seattle-cameras",
        "sd": 2973238,
        "sst": 79,
        "template": "template-2",
        "traffic-class": "class-2",
        "upf": "starbucks-seattle-cameras",
        "uplink": 5
      }
    ]
  }
}`},
			}, nil
		}).AnyTimes()

	_, err := irange.GetDevice(gnmiClient)
	assert.NoError(t, err)
}

func TestImsiRange_CollapseImsi(t *testing.T) {

	dg1 := "dg-1"
	s1 := "site-1"
	e1 := "ent-1"
	e111 := uint32(111)
	mcc123 := uint32(123)
	mnc456 := uint32(456)
	format1 := "CCCNNNEEESSSSSS"
	dg := &models.Device{
		DeviceGroup: &models.DeviceGroup_DeviceGroup{
			DeviceGroup: map[string]*models.DeviceGroup_DeviceGroup_DeviceGroup{
				dg1: {
					Id:   &dg1,
					Site: &s1,
				},
			},
		},
		Site: &models.Site_Site{
			Site: map[string]*models.Site_Site_Site{
				s1: {
					Id:         &s1,
					Enterprise: &e1,
					ImsiDefinition: &models.Site_Site_Site_ImsiDefinition{
						Enterprise: &e111,
						Format:     &format1,
						Mcc:        &mnc456,
						Mnc:        &mcc123,
					},
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	gnmiClient := mocks.NewMockGnmiInterface(ctrl)

	err := irange.CollapseImsi(dg, gnmiClient)
	assert.NoError(t, err)

}

func TestImsiRange_AddImsiRange(t *testing.T) {
	dg1 := "dg-1"
	s1 := "site-1"
	e1 := "ent-1"
	e111 := uint32(111)
	mcc123 := uint32(123)
	mnc456 := uint32(456)
	format1 := "CCCNNNEEESSSSSS"
	dg := &models.Device{
		DeviceGroup: &models.DeviceGroup_DeviceGroup{
			DeviceGroup: map[string]*models.DeviceGroup_DeviceGroup_DeviceGroup{
				dg1: {
					Id:   &dg1,
					Site: &s1,
				},
			},
		},
		Site: &models.Site_Site{
			Site: map[string]*models.Site_Site_Site{
				s1: {
					Id:         &s1,
					Enterprise: &e1,
					ImsiDefinition: &models.Site_Site_Site_ImsiDefinition{
						Enterprise: &e111,
						Format:     &format1,
						Mcc:        &mnc456,
						Mnc:        &mcc123,
					},
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	gnmiClient := mocks.NewMockGnmiInterface(ctrl)

	gnmiClient.EXPECT().Update(gomock.Any(), gomock.Any(), "connectivity-service-v3",
		"onos-config.micro-onos.svc.cluster.local:5150", gomock.Any()).AnyTimes()
	firstimsi := uint64(123456111000001)
	lastimsi := uint64(123456111000005)
	err := irange.AddImsiRange(dg, gnmiClient, "dg-1", firstimsi, lastimsi)
	assert.NoError(t, err)
}

func TestImsiRange_DeleteImsiRanges(t *testing.T) {
	dg1 := "dg-1"
	s1 := "site-1"
	e1 := "ent-1"
	e111 := uint32(111)
	mcc123 := uint32(123)
	mnc456 := uint32(456)
	format1 := "CCCNNNEEESSSSSS"
	dg := &models.Device{
		DeviceGroup: &models.DeviceGroup_DeviceGroup{
			DeviceGroup: map[string]*models.DeviceGroup_DeviceGroup_DeviceGroup{
				dg1: {
					Id:   &dg1,
					Site: &s1,
				},
			},
		},
		Site: &models.Site_Site{
			Site: map[string]*models.Site_Site_Site{
				s1: {
					Id:         &s1,
					Enterprise: &e1,
					ImsiDefinition: &models.Site_Site_Site_ImsiDefinition{
						Enterprise: &e111,
						Format:     &format1,
						Mcc:        &mnc456,
						Mnc:        &mcc123,
					},
				},
			},
		},
	}

	ctrl := gomock.NewController(t)
	gnmiClient := mocks.NewMockGnmiInterface(ctrl)

	gnmiClient.EXPECT().Delete(gomock.Any(), gomock.Any(), "connectivity-service-v3",
		"onos-config.micro-onos.svc.cluster.local:5150", gomock.Any()).AnyTimes()

	dgID := "dg-1"
	imsi := Imsi{
		123456111000001, 12345611100005, "auto-12345611100001"}

	err := irange.DeleteImsiRanges(dg, gnmiClient, dgID, imsi)
	assert.NoError(t, err)
}

func Test_mergeImsi(t *testing.T) {

	var imsi = []Imsi{{111222333000001, 111222333000003, "auto-111222333000001"},
		{111222333000005, 111222333000006, "auto-111222333000005"},
		{111222333000007, 111222333000008, "auto-111222333000007"},
	}

	newRange, delRanges := mergeImsi(imsi)
	if newRange[0].firstImsi != 111222333000001 || newRange[0].lastImsi != 111222333000003 {
		t.Errorf("need firstImsi as 111222333000001 and lastImsi as 111222333000003 but got %d and %d",
			newRange[0].firstImsi, newRange[0].lastImsi)
	}

	if newRange[1].firstImsi != 111222333000005 || newRange[1].lastImsi != 111222333000008 {
		t.Errorf("need firstImsi as 111222333000005 and lastImsi as 111222333000008 but got %d and %d",
			newRange[1].firstImsi, newRange[1].lastImsi)
	}

	for im := range delRanges {
		if !(im == "auto-111222333000005" || im == "auto-111222333000007") {
			t.Errorf("need delete imsi as auto-111222333000005 or auto-111222333000007 but got %s", im)
		}
	}
}
