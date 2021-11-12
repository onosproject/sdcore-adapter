// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package synchronizer

import (
	"github.com/golang/mock/gomock"
	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
	"github.com/onosproject/sdcore-adapter/pkg/test/mocks"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
	"testing"
)

func BuildRootPath(model string, key string) *pb.Path {
	keyMap := map[string]string{"id": key}
	path := &pb.Path{Elem: []*pb.PathElem{{Name: model}, {Name: model, Key: keyMap}}}
	return path
}

// Test cases where HandleDevice does nothing
func TestHandleDeleteNotApplicable(t *testing.T) {
	s := Synchronizer{}

	// Path is nil
	device := &models.Device{}
	err := s.HandleDelete(device, nil)
	assert.Nil(t, err)

	// Path has no elements
	path := &pb.Path{}
	err = s.HandleDelete(device, path)
	assert.Nil(t, err)

	// Path has only one element
	path = &pb.Path{Elem: []*pb.PathElem{{Name: "anything"}}}
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Refusing to delete path anything because it is too broad")

	// Path is not for a vcs or device-group
	path = &pb.Path{Elem: []*pb.PathElem{{Name: "anything"}, {Name: "else"}}}
	err = s.HandleDelete(device, path)
	assert.Nil(t, err)

	// Path is for a vcs but lacks a key
	path = &pb.Path{Elem: []*pb.PathElem{{Name: "vcs"}, {Name: "vcs"}}}
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of vcs does not have an id key")

	// Path is for a device-group but lacks a key
	path = &pb.Path{Elem: []*pb.PathElem{{Name: "device-group"}, {Name: "device-group"}}}
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of device-group does not have an id key")

	// Path is for a leaf within a vcs
	path = BuildRootPath("vcs", "sample-vcs")
	path.Elem = append(path.Elem, &pb.PathElem{Name: "leaf"})
	err = s.HandleDelete(device, path)
	assert.Nil(t, err)

	// Path is for a leaf within a device-group
	path = BuildRootPath("device-group", "sample-vcs")
	path.Elem = append(path.Elem, &pb.PathElem{Name: "sample-dg"})
	err = s.HandleDelete(device, path)
	assert.Nil(t, err)
}

func TestHandleDeleteVcs(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

	device := BuildSampleDevice()
	path := BuildRootPath("vcs", "sample-vcs")
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/network-slice/sample-vcs").DoAndReturn(func(endpoint string) error {
		return nil
	}).AnyTimes()
	err := s.HandleDelete(device, path)
	assert.Nil(t, err)
}

func TestHandleDeleteVcsPushError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

	// 404 is treated as a non-error, because we may have already deleted it
	device := BuildSampleDevice()
	path := BuildRootPath("vcs", "sample-vcs")
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/network-slice/sample-vcs").DoAndReturn(func(endpoint string) error {
		return &PushError{Operation: "DELETE", Endpoint: endpoint, StatusCode: 404, Status: "Not Found"}
	}).AnyTimes()
	err := s.HandleDelete(device, path)
	assert.Nil(t, err)

	// reset the mockpusher between tests
	mockPusher = mocks.NewMockPusherInterface(ctrl)
	s.SetPusher(mockPusher)

	// 403 is a problem
	device = BuildSampleDevice()
	path = BuildRootPath("vcs", "sample-vcs")
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/network-slice/sample-vcs").DoAndReturn(func(endpoint string) error {
		return &PushError{Operation: "DELETE", Endpoint: endpoint, StatusCode: 403, Status: "Forbidden"}
	}).AnyTimes()
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Vcs sample-vcs failed to push delete: Push Error op=DELETE endpoint=http://5gcore/v1/network-slice/sample-vcs code=403 status=Forbidden")
}

func TestHandleDeleteVCSMissingDeps(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

	// Vcs is nil
	device := BuildSampleDevice()
	device.Vcs = nil
	path := BuildRootPath("vcs", "sample-vcs")
	err := s.HandleDelete(device, path)
	assert.EqualError(t, err, "No VCSes")

	// Vcs is empty
	device = BuildSampleDevice()
	device.Vcs = &models.OnfVcs_Vcs{}
	path = BuildRootPath("vcs", "sample-vcs")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Vcs sample-vcs not found")

	// Site is nil
	device = BuildSampleDevice()
	device.Site = nil
	path = BuildRootPath("vcs", "sample-vcs")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "No Sites")

	// Site is empty list
	device = BuildSampleDevice()
	device.Site = &models.OnfSite_Site{}
	path = BuildRootPath("vcs", "sample-vcs")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Site sample-site not found")

	// Enterprise is nil
	device = BuildSampleDevice()
	device.Enterprise = nil
	path = BuildRootPath("vcs", "sample-vcs")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "No Enterprises")

	// Enterprise is empty list
	device = BuildSampleDevice()
	device.Enterprise = &models.OnfEnterprise_Enterprise{}
	path = BuildRootPath("vcs", "sample-vcs")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Enterprise sample-ent not found")

	// ConnectivityService is nil
	device = BuildSampleDevice()
	device.ConnectivityService = nil
	path = BuildRootPath("vcs", "sample-vcs")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "No connectivity services")

	// ConnectivityService is empty list
	device = BuildSampleDevice()
	device.ConnectivityService = &models.OnfConnectivityService_ConnectivityService{}
	path = BuildRootPath("vcs", "sample-vcs")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "ConnectivityService sample-cs not found")
}

func TestHandleDeleteDeviceGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

	device := BuildSampleDevice()
	path := BuildRootPath("device-group", "sample-dg")
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/device-group/sample-dg").DoAndReturn(func(endpoint string) error {
		return nil
	}).AnyTimes()
	err := s.HandleDelete(device, path)
	assert.Nil(t, err)
}

func TestHandleDeleteDeviceGroupPushError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

	// 404 is treated as a non-error, because we may have already deleted it
	device := BuildSampleDevice()
	path := BuildRootPath("device-group", "sample-dg")
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/device-group/sample-dg").DoAndReturn(func(endpoint string) error {
		return &PushError{Operation: "DELETE", Endpoint: endpoint, StatusCode: 404, Status: "Not Found"}
	}).AnyTimes()
	err := s.HandleDelete(device, path)
	assert.Nil(t, err)

	// reset the mockpusher between tests
	mockPusher = mocks.NewMockPusherInterface(ctrl)
	s.SetPusher(mockPusher)

	// 403 is a problem
	device = BuildSampleDevice()
	path = BuildRootPath("device-group", "sample-dg")
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/device-group/sample-dg").DoAndReturn(func(endpoint string) error {
		return &PushError{Operation: "DELETE", Endpoint: endpoint, StatusCode: 403, Status: "Forbidden"}
	}).AnyTimes()
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Device-Group sample-dg failed to push delete: Push Error op=DELETE endpoint=http://5gcore/v1/device-group/sample-dg code=403 status=Forbidden")
}

func TestHandleDeleteDeviceGroupMissingDeps(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := Synchronizer{}
	s.SetPusher(mockPusher)

	// DeviceGroup is nil
	device := BuildSampleDevice()
	device.DeviceGroup = nil
	path := BuildRootPath("device-group", "sample-dg")
	err := s.HandleDelete(device, path)
	assert.EqualError(t, err, "No DeviceGroups")

	// Vcs is empty
	device = BuildSampleDevice()
	device.DeviceGroup = &models.OnfDeviceGroup_DeviceGroup{}
	path = BuildRootPath("device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "DeviceGroup sample-dg not found")

	// Site is nil
	device = BuildSampleDevice()
	device.Site = nil
	path = BuildRootPath("device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "No Sites")

	// Site is empty list
	device = BuildSampleDevice()
	device.Site = &models.OnfSite_Site{}
	path = BuildRootPath("device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Site sample-site not found")

	// Enterprise is nil
	device = BuildSampleDevice()
	device.Enterprise = nil
	path = BuildRootPath("device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "No Enterprises")

	// Enterprise is empty list
	device = BuildSampleDevice()
	device.Enterprise = &models.OnfEnterprise_Enterprise{}
	path = BuildRootPath("device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Enterprise sample-ent not found")

	// ConnectivityService is nil
	device = BuildSampleDevice()
	device.ConnectivityService = nil
	path = BuildRootPath("device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "No connectivity services")

	// ConnectivityService is empty list
	device = BuildSampleDevice()
	device.ConnectivityService = &models.OnfConnectivityService_ConnectivityService{}
	path = BuildRootPath("device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "ConnectivityService sample-cs not found")
}
