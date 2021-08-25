// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package migration

import (
	"context"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/onosproject/sdcore-adapter/pkg/test/mocks"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

// Deprecated. Use local variable instead
var V1SetRequests []*gpb.SetRequest

// Deprecated. Use local variable instead
var V2SetRequests []*gpb.SetRequest

// Mock a gNMI Get Request, providing a mocked v1 device.
// Deprecated. Use local EXPECT() instead
func MigrationTestMockGet(req *gpb.GetRequest) (*gpb.GetResponse, error) {
	if len(req.Path) == 0 {
		return nil, errors.New("Get: No Path")
	}
	if req.Path[0].Target == "v1-device" {
		// construct an update, notification, and GetResponse
		jsonStr := "{}"
		update := gpb.Update{Path: req.Path[0],
			Val: &gpb.TypedValue{Value: &gpb.TypedValue_StringVal{StringVal: jsonStr}}}
		notification := gpb.Notification{Update: []*gpb.Update{&update}}
		return &gpb.GetResponse{Notification: []*gpb.Notification{&notification}}, nil
	} else if req.Path[0].Target == "v2-device" {
		return &gpb.GetResponse{}, nil
	}
	return nil, errors.New("Get: Unknown target")
}

// Mock a gNMI Set Reqeust, storing the sets in V1SetRequests or V2SetRequests
// for further examination.
// Deprecated. Use local EXPECT() instead
func MigrationTestMockSet(req *gpb.SetRequest) (*gpb.SetResponse, error) {
	var path *gpb.Path
	if req.Prefix != nil {
		path = req.Prefix
	} else if req.Update != nil {
		path = req.Update[0].Path
	} else if req.Delete != nil {
		path = req.Delete[0]
	} else {
		return nil, errors.New("Set: No Prefix or Update Path or Delete Path")
	}
	if path.Target == "v1-device" {
		V1SetRequests = append(V1SetRequests, req)
		return &gpb.SetResponse{}, nil
	} else if path.Target == "v2-device" {
		V2SetRequests = append(V2SetRequests, req)
		return &gpb.SetResponse{}, nil
	}
	return nil, errors.New("Set: Unknown target")
}

// Create a mock action that updates a leaf, and then deletes the source model.
func MakeMockAction(fromTarget string, toTarget string, updatePrefixStr string, updatePathStr string, val string) *MigrationActions {
	updatePrefix := StringToPath(updatePrefixStr, toTarget)
	update := UpdateString(updatePathStr, toTarget, &val)
	updates := []*gpb.Update{update}
	deletePath := StringToPath(updatePrefixStr, fromTarget)
	deletes := []*gpb.Path{deletePath}
	return &MigrationActions{UpdatePrefix: updatePrefix, Updates: updates, Deletes: deletes}
}

// Mock migration step from V1 to V2 model.
func MigrateV1V2(step *MigrationStep, fromTarget string, toTarget string, srcVal *gpb.TypedValue, destVal *gpb.TypedValue) ([]*MigrationActions, error) {
	action := MakeMockAction(fromTarget, toTarget, "/prefix", "/path/to/name", "value")
	return []*MigrationActions{action}, nil
}

// Mock migration step from V2 to V3 model.
func MigrateV2V3(step *MigrationStep, fromTarget string, toTarget string, srcVal *gpb.TypedValue, destVal *gpb.TypedValue) ([]*MigrationActions, error) {
	return []*MigrationActions{}, nil
}

// Mock migration step from V3 to V4 model.
func MigrateV3V4(step *MigrationStep, fromTarget string, toTarget string, srcVal *gpb.TypedValue, destVal *gpb.TypedValue) ([]*MigrationActions, error) {
	return []*MigrationActions{}, nil
}

// Mock migration step from V5 to V6 model.
func MigrateV5V6(step *MigrationStep, fromTarget string, toTarget string, srcVal *gpb.TypedValue, destVal *gpb.TypedValue) ([]*MigrationActions, error) {
	return []*MigrationActions{}, nil
}

func TestAddMigrationStep(t *testing.T) {
	ctrl := gomock.NewController(t)
	gnmiClient := mocks.NewMockGnmiInterface(ctrl)

	m := NewMigrator(gnmiClient)

	v1Models := &gnmi.Model{}
	v2Models := &gnmi.Model{}

	assert.Empty(t, m.steps)

	m.AddMigrationStep("1.0.0", v1Models, "2.0.0", v2Models, MigrateV1V2)

	assert.Len(t, m.steps, 1)

	assert.Equal(t, "1.0.0", m.steps[0].FromVersion)
	assert.Equal(t, v1Models, m.steps[0].FromModels)
	assert.Equal(t, "2.0.0", m.steps[0].ToVersion)
	assert.Equal(t, v2Models, m.steps[0].ToModels)
	assert.NotNil(t, m.steps[0].MigrationFunc)
	assert.Equal(t, m, m.steps[0].Migrator)
}

func TestBuildStepList(t *testing.T) {
	m := NewMigrator(nil)

	v1Models := &gnmi.Model{}
	v2Models := &gnmi.Model{}

	m.AddMigrationStep("1.0.0", v1Models, "2.0.0", v2Models, MigrateV1V2)
	m.AddMigrationStep("2.0.0", v1Models, "3.0.0", v2Models, MigrateV2V3)
	m.AddMigrationStep("3.0.0", v1Models, "4.0.0", v2Models, MigrateV3V4)
	m.AddMigrationStep("5.0.0", v1Models, "6.0.0", v2Models, MigrateV5V6)

	// transitive closure of three steps
	stepList, err := m.buildStepList("1.0.0", "4.0.0")
	assert.Nil(t, err)
	assert.Len(t, stepList, 3)

	// transitive closure of two steps
	stepList, err = m.buildStepList("1.0.0", "3.0.0")
	assert.Nil(t, err)
	assert.Len(t, stepList, 2)

	// starting in the middle
	stepList, err = m.buildStepList("2.0.0", "4.0.0")
	assert.Nil(t, err)
	assert.Len(t, stepList, 2)

	// the first version doesn't exist
	_, err = m.buildStepList("1.0.11", "2.0.0")
	assert.EqualError(t, err, "Unable to find a step that started with version 1.0.11")

	// the last version doesn't exist
	_, err = m.buildStepList("1.0.0", "2.0.22")
	assert.EqualError(t, err, "Unable to find a step that ended with version 2.0.22")

	// transitive closure has a hole
	_, err = m.buildStepList("1.0.0", "6.0.0")
	assert.EqualError(t, err, "Unable to find a step that ended with version 6.0.0")
}

func TestRunStep(t *testing.T) {
	ctrl := gomock.NewController(t)
	gnmiClient := mocks.NewMockGnmiInterface(ctrl)
	gnmiClient.EXPECT().Address().Return("testaddress").Times(2)
	gnmiClient.EXPECT().GetPath(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, path string, target string, addr string) (*gpb.TypedValue, error) {
			return &gpb.TypedValue{
				Value: &gpb.TypedValue_StringVal{StringVal: "{}"},
			}, nil
		}).Times(2)
	m := NewMigrator(gnmiClient)

	v1Models := &gnmi.Model{}
	v2Models := &gnmi.Model{}

	m.AddMigrationStep("1.0.0", v1Models, "2.0.0", v2Models, MigrateV1V2)
	// TODO: We're going to need to mock the gNMI get and set

	actions, err := m.runStep(m.steps[0], "v1-device", "v2-device")
	assert.Nil(t, err)

	// The step should have added an update request
	assert.Len(t, actions, 1)
	assert.Equal(t, "v2-device", actions[0].UpdatePrefix.Target)
}

func TestExecuteActions(t *testing.T) {
	ctrl := gomock.NewController(t)
	gnmiClient := mocks.NewMockGnmiInterface(ctrl)
	gnmiClient.EXPECT().Address().Return("testaddress").AnyTimes()
	// Setup the mocks for gNMI get and set
	var v1SetRequests []*gpb.SetRequest
	var v2SetRequests []*gpb.SetRequest

	gnmiClient.EXPECT().Update(gomock.Any(), gomock.Any(), "v2-device", gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, prefix *gpb.Path, target string, addr string, updates []*gpb.Update) error {
			v2SetRequests = append(v2SetRequests, &gpb.SetRequest{
				Update: updates,
			})
			return nil
		})
	gnmiClient.EXPECT().Delete(gomock.Any(), gomock.Any(), "v1-device", gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, prefix *gpb.Path, target string, addr string, deletes []*gpb.Path) error {
			v1SetRequests = append(v1SetRequests, &gpb.SetRequest{Delete: deletes})
			return nil
		})
	m := NewMigrator(gnmiClient)

	updatePrefix := StringToPath("/root", "v2-device")
	val := "somevalue"
	update := UpdateString("/path/to/leaf", "v2-device", &val)
	updates := []*gpb.Update{update}

	deletePath := StringToPath("/root", "v1-device")

	action := &MigrationActions{UpdatePrefix: updatePrefix, Updates: updates, Deletes: []*gpb.Path{deletePath}}
	actions := []*MigrationActions{action}

	err := m.executeActions(actions, "v1-device", "v2-device")
	assert.Nil(t, err)

	// one delete of the v1 model
	assert.Len(t, v1SetRequests, 1)
	assert.Len(t, v1SetRequests[0].GetDelete(), 1)
	assert.Len(t, v1SetRequests[0].GetUpdate(), 0)
	assert.Equal(t, `elem:{name:"root"} target:"v1-device"`,
		strings.ReplaceAll(v1SetRequests[0].GetDelete()[0].String(), "  ", " "))

	// one create of the v2 model
	assert.Len(t, v2SetRequests, 1)
	assert.Len(t, v2SetRequests[0].GetDelete(), 0)
	assert.Len(t, v2SetRequests[0].GetUpdate(), 1)
	assert.Equal(t, `path:{elem:{name:"path"} elem:{name:"to"} elem:{name:"leaf"} target:"v2-device"} val:{string_val:"somevalue"}`,
		strings.ReplaceAll(v2SetRequests[0].GetUpdate()[0].String(), "  ", " "))
}

func TestMigrate(t *testing.T) {
	ctrl := gomock.NewController(t)
	gnmiClient := mocks.NewMockGnmiInterface(ctrl)
	gnmiClient.EXPECT().Address().Return("testaddress").AnyTimes()
	gnmiClient.EXPECT().GetPath(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, path string, target string, addr string) (*gpb.TypedValue, error) {
			return &gpb.TypedValue{
				Value: &gpb.TypedValue_StringVal{StringVal: "{}"},
			}, nil
		}).Times(2)
	// Setup the mocks for gNMI get and set
	var v1SetRequests []*gpb.SetRequest
	var v2SetRequests []*gpb.SetRequest
	gnmiClient.EXPECT().Update(gomock.Any(), gomock.Any(), "v2-device", gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, prefix *gpb.Path, target string, addr string, updates []*gpb.Update) error {
			v2SetRequests = append(v2SetRequests, &gpb.SetRequest{
				Update: updates,
			})
			return nil
		})
	gnmiClient.EXPECT().Delete(gomock.Any(), gomock.Any(), "v1-device", gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, prefix *gpb.Path, target string, addr string, deletes []*gpb.Path) error {
			v1SetRequests = append(v1SetRequests, &gpb.SetRequest{Delete: deletes})
			return nil
		})
	m := NewMigrator(gnmiClient)

	v1Models := &gnmi.Model{}
	v2Models := &gnmi.Model{}

	m.AddMigrationStep("1.0.0", v1Models, "2.0.0", v2Models, MigrateV1V2)

	// Should cause the V1->V2 migration step to be executed.
	err := m.Migrate("v1-device", "1.0.0", "v2-device", "2.0.0")
	assert.Nil(t, err)

	// one delete of the v1 model
	assert.Len(t, v1SetRequests, 1)
	assert.Len(t, v1SetRequests[0].GetDelete(), 1)
	assert.Len(t, v1SetRequests[0].GetUpdate(), 0)
	assert.Equal(t, `elem:{name:"prefix"} target:"v1-device"`,
		strings.ReplaceAll(v1SetRequests[0].GetDelete()[0].String(), "  ", " "))

	// one create of the v2 model
	assert.Len(t, v2SetRequests, 1)
	assert.Len(t, v2SetRequests[0].GetDelete(), 0)
	assert.Len(t, v2SetRequests[0].GetUpdate(), 1)
	assert.Equal(t, `path:{elem:{name:"path"} elem:{name:"to"} elem:{name:"name"} target:"v2-device"} val:{string_val:"value"}`,
		strings.ReplaceAll(v2SetRequests[0].GetUpdate()[0].String(), "  ", " "))
}

func TestNewMigrator(t *testing.T) {
	ctrl := gomock.NewController(t)
	gnmiClient := mocks.NewMockGnmiInterface(ctrl)
	gnmiClient.EXPECT().Address().Return("aether-config.aether.org").AnyTimes()

	m := NewMigrator(gnmiClient)
	assert.NotNil(t, m)
	assert.Equal(t, "aether-config.aether.org", m.Gnmi.Address())
}
