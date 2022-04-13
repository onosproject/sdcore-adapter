// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package synchronizer

import (
	"github.com/golang/mock/gomock"
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

	config, device := BuildSampleConfig()
	ipd := device.Site["sample-site"].IpDomain["sample-ipd"]

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

	mockPusher.EXPECT().PushUpdate("http://5gcore/v1/network-slice/sample-slice", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		return nil
	}).AnyTimes()

	mockPusher.EXPECT().PushUpdate("http://upf/v1/config/network-slices", gomock.Any()).DoAndReturn(func(endpoint string, data []byte) error {
		return nil
	}).AnyTimes()

	pushErrors, err := s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
	assert.Equal(t, len(pushes), 1)
	require.JSONEq(t, jsonData, pushes[0])

	// push it again, should not be any new pushes
	pushErrors, err = s.SynchronizeDevice(config)
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
	pushErrors, err = s.SynchronizeDevice(config)
	assert.Equal(t, 0, pushErrors)
	assert.Nil(t, err)
	assert.Equal(t, len(pushes), 2)
	require.JSONEq(t, jsonData, pushes[0])
	require.JSONEq(t, jsonDataUpdated, pushes[1])
}
