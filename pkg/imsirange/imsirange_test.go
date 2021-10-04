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
	dgJSON, err := ioutil.ReadFile("./testdata/deviceGroup.json")
	assert.NoError(t, err)

	siteJSON, err := ioutil.ReadFile("./testdata/deviceSite.json")
	assert.NoError(t, err)

	ctrl := gomock.NewController(t)
	gnmiClient := mocks.NewMockGnmiInterface(ctrl)
	gnmiClient.EXPECT().GetPath(gomock.Any(), "/device-group", "connectivity-service-v3",
		"onos-config.micro-onos.svc.cluster.local:5150").
		DoAndReturn(func(ctx context.Context, path string, target string, addr string) (*gpb.TypedValue, error) {
			return &gpb.TypedValue{
				Value: &gpb.TypedValue_JsonVal{JsonVal: dgJSON},
			}, nil
		}).AnyTimes()

	gnmiClient.EXPECT().GetPath(gomock.Any(), "/site", "connectivity-service-v3",
		"onos-config.micro-onos.svc.cluster.local:5150").
		DoAndReturn(func(ctx context.Context, path string, target string, addr string) (*gpb.TypedValue, error) {
			return &gpb.TypedValue{
				Value: &gpb.TypedValue_JsonVal{JsonVal: siteJSON},
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

	dgJSON, err := ioutil.ReadFile("./testdata/deviceGroup.json")
	assert.NoError(t, err)

	siteJSON, err := ioutil.ReadFile("./testdata/deviceSite.json")
	assert.NoError(t, err)

	device := &models.Device{}
	if len(dgJSON) > 0 {
		if err := models.Unmarshal(dgJSON, device); err != nil {
			assert.Error(t, err)
		}
	}
	if len(siteJSON) > 0 {
		if err := models.Unmarshal(siteJSON, device); err != nil {
			assert.Error(t, err)
		}
	}
	ctrl := gomock.NewController(t)
	gnmiClient := mocks.NewMockGnmiInterface(ctrl)

	var delSetRequests []*gpb.SetRequest
	var updSetRequests []*gpb.SetRequest

	gnmiClient.EXPECT().Delete(gomock.Any(), gomock.Any(), "connectivity-service-v3",
		"onos-config.micro-onos.svc.cluster.local:5150", gomock.Any()).
		DoAndReturn(func(ctx context.Context, prefix *gpb.Path, target string, addr string, deletes []*gpb.Path) error {
			delSetRequests = append(delSetRequests, &gpb.SetRequest{Delete: deletes, Prefix: prefix})
			return nil
		}).AnyTimes()

	gnmiClient.EXPECT().Update(gomock.Any(), gomock.Any(), "connectivity-service-v3",
		"onos-config.micro-onos.svc.cluster.local:5150", gomock.Any()).
		DoAndReturn(func(ctx context.Context, prefix *gpb.Path, target string, addr string, updates []*gpb.Update) error {
			updSetRequests = append(updSetRequests, &gpb.SetRequest{
				Update: updates,
			})
			return nil
		}).AnyTimes()

	err = irange.CollapseImsi(device, gnmiClient)

	assert.NotNil(t, delSetRequests)
	assert.Len(t, delSetRequests, 4)
	assert.Len(t, delSetRequests[3].GetDelete(), 2)
	assert.Len(t, delSetRequests[0].Prefix.GetElem(), 3)
	for _, data := range delSetRequests {
		switch data.Prefix.Elem[2].Key["name"] {
		case "auto-111222333000010", "auto-111222333000015", "auto-21322-91", "auto-21032002000094":
		default:
			t.Errorf("unexpected imsi %v", data.Prefix.Elem[2].Key["name"])
		}
	}

	assert.NotNil(t, updSetRequests)
	assert.Len(t, updSetRequests, 2)
	assert.Len(t, updSetRequests[0].GetUpdate(), 2)
	for _, data := range updSetRequests {
		switch data.Update[0].Val.GetUintVal() {
		case 111222333000010, 91:
		default:
			t.Errorf("unexpected imsi %v", data.Update[0].Val.GetUintVal())
		}
	}

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
