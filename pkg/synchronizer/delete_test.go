// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package synchronizer

import (
	"github.com/golang/mock/gomock"
	models "github.com/onosproject/aether-models/models/aether-2.0.x/api"
	"github.com/onosproject/sdcore-adapter/pkg/test/mocks"
	pb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
	"testing"
)

// BuildRootPath builds a root path to a slice or dg
func BuildRootPath(entID string, siteID string, modelKey string, model string, key string) *pb.Path {
	entKeyMap := map[string]string{"enterprise-id": entID}
	siteKeyMap := map[string]string{"site-id": siteID}
	keyMap := map[string]string{modelKey: key}
	path := &pb.Path{Elem: []*pb.PathElem{{Name: "enterprises"}, {Name: "enterprise", Key: entKeyMap}, {Name: "site", Key: siteKeyMap}, {Name: model, Key: keyMap}}}
	return path
}

// Test cases where HandleDevice does nothing
func TestHandleDeleteNotApplicable(t *testing.T) {
	s := NewSynchronizer()

	entKey := map[string]string{"enterprise-id": "sample-ent"}
	siteKey := map[string]string{"site-id": "sample-site"}

	device := BuildSampleDevice()

	// Path is nil
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

	// Path is not for a slice or device-group
	path = &pb.Path{Elem: []*pb.PathElem{{Name: "anything"}, {Name: "else"}, {Name: "at"}, {Name: "all"}}}
	err = s.HandleDelete(device, path)
	assert.Nil(t, err)

	// Path is for a slice but lacks an enterprise key
	path = &pb.Path{Elem: []*pb.PathElem{{Name: "enterprises"}, {Name: "enterprise"}, {Name: "site"}, {Name: "slice"}}}
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of slice does not have an enterprise-id key")

	// Path is for a slice but lacks a site key
	path = &pb.Path{Elem: []*pb.PathElem{{Name: "enterprises"}, {Name: "enterprise", Key: entKey}, {Name: "site"}, {Name: "slice"}}}
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of slice does not have a site-id key")

	// Path is for a slice but lacks a key
	path = &pb.Path{Elem: []*pb.PathElem{{Name: "enterprises"}, {Name: "enterprise", Key: entKey}, {Name: "site", Key: siteKey}, {Name: "slice"}}}
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of slice does not have an id key")

	// Path is for a device-group but lacks a key
	path = &pb.Path{Elem: []*pb.PathElem{{Name: "enterprises"}, {Name: "enterprise", Key: entKey}, {Name: "site", Key: siteKey}, {Name: "device-group"}}}
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of device-group does not have an id key")

	// Path is for a leaf within a slice
	path = &pb.Path{Elem: []*pb.PathElem{{Name: "enterprises"}, {Name: "enterprise", Key: entKey}, {Name: "site", Key: siteKey}, {Name: "slice"}, {Name: "inside"}}}
	path.Elem = append(path.Elem, &pb.PathElem{Name: "leaf"})
	err = s.HandleDelete(device, path)
	assert.Nil(t, err)

	// Path is for a leaf within a device-group
	path = &pb.Path{Elem: []*pb.PathElem{{Name: "enterprises"}, {Name: "enterprise", Key: entKey}, {Name: "site", Key: siteKey}, {Name: "device-group"}, {Name: "inside"}}}
	path.Elem = append(path.Elem, &pb.PathElem{Name: "sample-dg"})
	err = s.HandleDelete(device, path)
	assert.Nil(t, err)
}

func TestHandleDeleteSlice(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := NewSynchronizer(WithPusher(mockPusher))

	device := BuildSampleDevice()
	path := BuildRootPath("sample-ent", "sample-site", "slice-id", "slice", "sample-slice")
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/network-slice/sample-slice").DoAndReturn(func(endpoint string) error {
		return nil
	}).AnyTimes()
	err := s.HandleDelete(device, path)
	assert.Nil(t, err)
}

func TestHandleDeleteSlicePushError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := NewSynchronizer(WithPusher(mockPusher))

	// 404 is treated as a non-error, because we may have already deleted it
	device := BuildSampleDevice()
	path := BuildRootPath("sample-ent", "sample-site", "slice-id", "slice", "sample-slice")
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/network-slice/sample-slice").DoAndReturn(func(endpoint string) error {
		return &PushError{Operation: "DELETE", Endpoint: endpoint, StatusCode: 404, Status: "Not Found"}
	}).AnyTimes()
	err := s.HandleDelete(device, path)
	assert.Nil(t, err)

	// reset the mockpusher and synchronizer between tests
	mockPusher = mocks.NewMockPusherInterface(ctrl)
	s = NewSynchronizer(WithPusher(mockPusher))

	// 403 is a problem
	device = BuildSampleDevice()
	path = BuildRootPath("sample-ent", "sample-site", "slice-id", "slice", "sample-slice")
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/network-slice/sample-slice").DoAndReturn(func(endpoint string) error {
		return &PushError{Operation: "DELETE", Endpoint: endpoint, StatusCode: 403, Status: "Forbidden"}
	}).AnyTimes()
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Slice sample-slice failed to push delete: Push Error op=DELETE endpoint=http://5gcore/v1/network-slice/sample-slice code=403 status=Forbidden")
}

func TestHandleDeleteVCSMissingDeps(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := NewSynchronizer(WithPusher(mockPusher))

	// Slice is nil
	device := BuildSampleDevice()
	device.Enterprises.Enterprise["sample-ent"].Site["sample-site"].Slice = nil
	path := BuildRootPath("sample-ent", "sample-site", "slice-id", "slice", "sample-slice")
	err := s.HandleDelete(device, path)
	assert.EqualError(t, err, "Slice sample-slice not found")

	// Slice is empty
	device = BuildSampleDevice()
	device.Enterprises.Enterprise["sample-ent"].Site["sample-site"].Slice = map[string]*Slice{}
	path = BuildRootPath("sample-ent", "sample-site", "slice-id", "slice", "sample-slice")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Slice sample-slice not found")

	// Site is nil
	device = BuildSampleDevice()
	device.Enterprises.Enterprise["sample-ent"].Site = nil
	path = BuildRootPath("sample-ent", "sample-site", "slice-id", "slice", "sample-slice")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of slice failed to find site sample-site")

	// Site is empty list
	device = BuildSampleDevice()
	device.Enterprises.Enterprise["sample-ent"].Site = map[string]*Site{}
	path = BuildRootPath("sample-ent", "sample-site", "slice-id", "slice", "sample-slice")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of slice failed to find site sample-site")

	// Enterprises is nil
	device = BuildSampleDevice()
	device.Enterprises = nil
	path = BuildRootPath("sample-ent", "sample-site", "slice-id", "slice", "sample-slice")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of slice failed to find enterprise sample-ent")

	// Enterprise is empty list
	device = BuildSampleDevice()
	device.Enterprises = &models.OnfEnterprise_Enterprises{}
	path = BuildRootPath("sample-ent", "sample-site", "slice-id", "slice", "sample-slice")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of slice failed to find enterprise sample-ent")

	// ConnectivityServices is nil
	device = BuildSampleDevice()
	device.ConnectivityServices = nil
	path = BuildRootPath("sample-ent", "sample-site", "slice-id", "slice", "sample-slice")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "No connectivity services")

	// ConnectivityService is empty list
	device = BuildSampleDevice()
	device.ConnectivityServices = &models.OnfConnectivityService_ConnectivityServices{}
	path = BuildRootPath("sample-ent", "sample-site", "slice-id", "slice", "sample-slice")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "No connectivity services")
}

func TestHandleDeleteDeviceGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := NewSynchronizer(WithPusher(mockPusher))

	device := BuildSampleDevice()
	path := BuildRootPath("sample-ent", "sample-site", "dg-id", "device-group", "sample-dg")
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/device-group/sample-dg").DoAndReturn(func(endpoint string) error {
		return nil
	}).AnyTimes()
	err := s.HandleDelete(device, path)
	assert.Nil(t, err)
}

func TestHandleDeleteDeviceGroupPushError(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := NewSynchronizer(WithPusher(mockPusher))

	// 404 is treated as a non-error, because we may have already deleted it
	device := BuildSampleDevice()
	path := BuildRootPath("sample-ent", "sample-site", "dg-id", "device-group", "sample-dg")
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/device-group/sample-dg").DoAndReturn(func(endpoint string) error {
		return &PushError{Operation: "DELETE", Endpoint: endpoint, StatusCode: 404, Status: "Not Found"}
	}).AnyTimes()
	err := s.HandleDelete(device, path)
	assert.Nil(t, err)

	// reset the mockpusher and synchronizer between tests
	mockPusher = mocks.NewMockPusherInterface(ctrl)
	s = NewSynchronizer(WithPusher(mockPusher))

	// 403 is a problem
	device = BuildSampleDevice()
	path = BuildRootPath("sample-ent", "sample-site", "dg-id", "device-group", "sample-dg")
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/device-group/sample-dg").DoAndReturn(func(endpoint string) error {
		return &PushError{Operation: "DELETE", Endpoint: endpoint, StatusCode: 403, Status: "Forbidden"}
	}).AnyTimes()
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Device-Group sample-dg failed to push delete: Push Error op=DELETE endpoint=http://5gcore/v1/device-group/sample-dg code=403 status=Forbidden")
}

func TestHandleDeleteDeviceGroupMissingDeps(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := NewSynchronizer(WithPusher(mockPusher))

	// DeviceGroup is nil
	device := BuildSampleDevice()
	device.Enterprises.Enterprise["sample-ent"].Site["sample-site"].DeviceGroup = nil
	path := BuildRootPath("sample-ent", "sample-site", "dg-id", "device-group", "sample-dg")
	err := s.HandleDelete(device, path)
	assert.EqualError(t, err, "DeviceGroup sample-dg not found")

	// Slice is empty
	device = BuildSampleDevice()
	device.Enterprises.Enterprise["sample-ent"].Site["sample-site"].DeviceGroup = map[string]*DeviceGroup{}
	path = BuildRootPath("sample-ent", "sample-site", "dg-id", "device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "DeviceGroup sample-dg not found")

	// Site is nil
	device = BuildSampleDevice()
	device.Enterprises.Enterprise["sample-ent"].Site = nil
	path = BuildRootPath("sample-ent", "sample-site", "dg-id", "device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of device-group failed to find site sample-site")

	// Site is empty list
	device = BuildSampleDevice()
	device.Enterprises.Enterprise["sample-ent"].Site = map[string]*Site{}
	path = BuildRootPath("sample-ent", "sample-site", "dg-id", "device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of device-group failed to find site sample-site")

	// Enterprises is nil
	device = BuildSampleDevice()
	device.Enterprises = nil
	path = BuildRootPath("sample-ent", "sample-site", "dg-id", "device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of device-group failed to find enterprise sample-ent")

	// Enterprises is empty list
	device = BuildSampleDevice()
	device.Enterprises = &models.OnfEnterprise_Enterprises{}
	path = BuildRootPath("sample-ent", "sample-site", "dg-id", "device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "Delete of device-group failed to find enterprise sample-ent")

	// ConnectivityService is nil
	device = BuildSampleDevice()
	device.ConnectivityServices = nil
	path = BuildRootPath("sample-ent", "sample-site", "dg-id", "device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "No connectivity services")

	// ConnectivityService is empty list
	device = BuildSampleDevice()
	device.ConnectivityServices = &models.OnfConnectivityService_ConnectivityServices{}
	path = BuildRootPath("sample-ent", "sample-site", "dg-id", "device-group", "sample-dg")
	err = s.HandleDelete(device, path)
	assert.EqualError(t, err, "No connectivity services")
}

func TestHandleDeleteSlite(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := NewSynchronizer(WithPusher(mockPusher))

	device := BuildSampleDevice()

	entKeyMap := map[string]string{"enterprise-id": "sample-ent"}
	siteKeyMap := map[string]string{"site-id": "sample-site"}
	path := &pb.Path{Elem: []*pb.PathElem{{Name: "enterprises"}, {Name: "enterprise", Key: entKeyMap}, {Name: "site", Key: siteKeyMap}}}
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/device-group/sample-dg").DoAndReturn(func(endpoint string) error {
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/network-slice/sample-slice").DoAndReturn(func(endpoint string) error {
		return nil
	}).AnyTimes()
	err := s.HandleDelete(device, path)
	assert.Nil(t, err)
}

func TestHandleDeleteEnterprise(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockPusher := mocks.NewMockPusherInterface(ctrl)
	s := NewSynchronizer(WithPusher(mockPusher))

	device := BuildSampleDevice()

	entKeyMap := map[string]string{"enterprise-id": "sample-ent"}
	path := &pb.Path{Elem: []*pb.PathElem{{Name: "enterprises"}, {Name: "enterprise", Key: entKeyMap}}}
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/device-group/sample-dg").DoAndReturn(func(endpoint string) error {
		return nil
	}).AnyTimes()
	mockPusher.EXPECT().PushDelete("http://5gcore/v1/network-slice/sample-slice").DoAndReturn(func(endpoint string) error {
		return nil
	}).AnyTimes()
	err := s.HandleDelete(device, path)
	assert.Nil(t, err)
}
