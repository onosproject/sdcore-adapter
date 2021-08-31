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
	"io/ioutil"
	"testing"
)

var irange = ImsiRange{
	AetherConfigAddress: "onos-config.micro-onos.svc.cluster.local:5150",
	AetherConfigTarget:  "connectivity-service-v3",
}

func TestImsiRange_GetDevice(t *testing.T) {
	dataJSON, err := ioutil.ReadFile("./testdata/deviceData.json")
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	gnmiClient := mocks.NewMockGnmiInterface(ctrl)
	gnmiClient.EXPECT().GetPath(gomock.Any(), "", "connectivity-service-v3",
		"onos-config.micro-onos.svc.cluster.local:5150").
		DoAndReturn(func(ctx context.Context, path string, target string, addr string) (*gpb.TypedValue, error) {
			return &gpb.TypedValue{
				Value: &gpb.TypedValue_JsonVal{JsonVal: dataJSON},
			}, nil
		}).AnyTimes()

	device, err := irange.GetDevice(gnmiClient)
	assert.NotNil(t, device.DeviceGroup)
	assert.Len(t, device.DeviceGroup.DeviceGroup, 9)
	assert.Equal(t, "Global Default Device Group",
		*device.DeviceGroup.DeviceGroup["defaultent-defaultsite-default"].DisplayName)
	assert.NotNil(t, device.Site)
	assert.Len(t, device.Site.Site, 4)
	assert.Equal(t, "New York", *device.Site.Site["starbucks-newyork"].DisplayName)
	assert.NoError(t, err)
}

func TestImsiRange_CollapseImsi(t *testing.T) {

	dataJSON, err := ioutil.ReadFile("./testdata/deviceData.json")
	assert.NoError(t, err)
	device := &models.Device{}
	if len(dataJSON) > 0 {
		if err := models.Unmarshal(dataJSON, device); err != nil {
			//return nil, errors.NewInvalid("Failed to unmarshal json")
			assert.Error(t, err)
		}
	}
	ctrl := gomock.NewController(t)
	gnmiClient := mocks.NewMockGnmiInterface(ctrl)

	gnmiClient.EXPECT().Delete(gomock.Any(), gomock.Any(), "connectivity-service-v3",
		"onos-config.micro-onos.svc.cluster.local:5150", gomock.Any()).AnyTimes()

	gnmiClient.EXPECT().Update(gomock.Any(), gomock.Any(), "connectivity-service-v3",
		"onos-config.micro-onos.svc.cluster.local:5150", gomock.Any()).AnyTimes()

	err = irange.CollapseImsi(device, gnmiClient)
	assert.NoError(t, err)

}

func Test_mergeImsi(t *testing.T) {

	var imsi = []Imsi{{111222333000001, 111222333000003, "auto-111222333000001"},
		{111222333000005, 111222333000006, "auto-111222333000005"},
		{111222333000007, 111222333000008, "auto-111222333000007"},
	}

	newRange, delRanges := mergeImsi(imsi)

	assert.Len(t, newRange, 2)

	assert.Equal(t, uint64(111222333000001), newRange[0].firstImsi)
	assert.Equal(t, uint64(111222333000003), newRange[0].lastImsi)
	assert.Equal(t, uint64(111222333000005), newRange[1].firstImsi)
	assert.Equal(t, uint64(111222333000008), newRange[1].lastImsi)

	assert.Len(t, delRanges, 2)
	for im := range delRanges {
		switch im {
		case "auto-111222333000005", "auto-111222333000007":
		default:
			t.Errorf("unexpected imsi %s", im)
		}
	}
}
