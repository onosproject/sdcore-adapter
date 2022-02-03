// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package synchronizer

import (
	"github.com/golang/mock/gomock"
	models "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
	"github.com/onosproject/sdcore-adapter/pkg/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSynchronizeDeviceDeviceGroupCacheTest(t *testing.T) {

	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	pushes := []string{}
	s := NewSynchronizer(WithPusher(mockPusher))

	ent, cs, _, ipd, _, _ := BuildSampleDeviceGroup() // nolint dogsled

	device := &RootDevice{
		Enterprises:          &models.OnfEnterprise_Enterprises{Enterprise: map[string]*Enterprise{"sample-ent": ent}},
		ConnectivityServices: &models.OnfConnectivityService_ConnectivityServices{ConnectivityService: map[string]*ConnectivityService{"sample-cs": cs}},
	}

	jsonData := `{
			"imsis": [
			  "123456789000001"
			],
			"ip-domain-name": "sample-ipd",
			"site-info": "sample-site",
			"ip-domain-expanded": {
			  "dnn": "5ginternet",
			  "ue-ip-pool": "1.2.3.4/24",
			  "dns-primary": "8.8.8.8",
				"mtu": 1492,
				"ue-dnn-qos": {
					"dnn-mbr-downlink": 4321,
					"dnn-mbr-uplink": 8765,
					"bitrate-unit": "bps",
					"traffic-class": {
						"name": "sample-traffic-class",
						"arp": 3,
						"pdb": 300,
						"pelr": 6,
						"qci": 55
					}					
				}				
			}
			}`
	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/device-group/sample-dg", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		pushes = append(pushes, string(data))
		return nil
	}).AnyTimes()

	pushErrors, err := s.SynchronizeDevice(device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
	assert.Equal(t, len(pushes), 1)
	require.JSONEq(t, jsonData, pushes[0])

	// push it again, should not be any new pushes
	pushErrors, err = s.SynchronizeDevice(device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
	assert.Equal(t, len(pushes), 1)
	require.JSONEq(t, jsonData, pushes[0])

	// Update the models
	ipd.DnsPrimary = aStr("5.6.7.8")

	jsonDataUpdated := `{
		"imsis": [
			"123456789000001"
		],
		"ip-domain-name": "sample-ipd",
		"site-info": "sample-site",
		"ip-domain-expanded": {
			"dnn": "5ginternet",
			"ue-ip-pool": "1.2.3.4/24",
			"dns-primary": "5.6.7.8",
			"mtu": 1492,
			"ue-dnn-qos": {
				"dnn-mbr-downlink": 4321,
				"dnn-mbr-uplink": 8765,
				"bitrate-unit": "bps",
				"traffic-class": {
					"name": "sample-traffic-class",
					"arp": 3,
					"pdb": 300,
					"pelr": 6,
					"qci": 55
				}					
			}				
		}
		}`

	// push it again, this time we should get a new push
	pushErrors, err = s.SynchronizeDevice(device)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
	assert.Equal(t, len(pushes), 2)
	require.JSONEq(t, jsonData, pushes[0])
	require.JSONEq(t, jsonDataUpdated, pushes[1])
}
