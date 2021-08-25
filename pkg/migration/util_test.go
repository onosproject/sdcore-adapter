// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package migration

import (
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitKey(t *testing.T) {
	// key and value
	name, key, value := splitKey("foo[a=b]")
	assert.Equal(t, *name, "foo")
	assert.Equal(t, *key, "a")
	assert.Equal(t, *value, "b")

	// no [] section
	name, key, value = splitKey("foo")
	assert.Equal(t, *name, "foo")
	assert.Nil(t, key)
	assert.Nil(t, value)

	// [] section is empty
	name, key, value = splitKey("foo[]")
	assert.Equal(t, *name, "foo")
	assert.Nil(t, key)
	assert.Nil(t, value)

	// has key, but no value
	// reasonable behavior is to treat the key=value as undefined and ignore it
	name, key, value = splitKey("foo[junk]")
	assert.Equal(t, *name, "foo")
	assert.Nil(t, key)
	assert.Nil(t, value)

	// empty string
	name, key, value = splitKey("")
	assert.Nil(t, name)
	assert.Nil(t, key)
	assert.Nil(t, value)

	// just a key=value but no name
	// reasonable behavior is to treat this the same as empty string
	name, key, value = splitKey("[foo=bar]")
	assert.Nil(t, name)
	assert.Nil(t, key)
	assert.Nil(t, value)
}

func TestStringToPath(t *testing.T) {
	p := StringToPath("/foo", "v1-device")
	assert.NotNil(t, p)
	assert.Equal(t, p.Target, "v1-device")
	assert.Len(t, p.Elem, 1)
	assert.Equal(t, p.Elem[0].Name, "foo")
	assert.Empty(t, p.Elem[0].Key)

	p = StringToPath("/foo[a=b]", "v1-device")
	assert.NotNil(t, p)
	assert.Equal(t, p.Target, "v1-device")
	assert.Len(t, p.Elem, 1)
	assert.Equal(t, p.Elem[0].Name, "foo")
	assert.Equal(t, map[string]string{"a": "b"}, p.Elem[0].Key)

	p = StringToPath("/one/two[a=b]/three", "v1-device")
	assert.NotNil(t, p)
	assert.Equal(t, p.Target, "v1-device")
	assert.Len(t, p.Elem, 3)
	assert.Equal(t, p.Elem[0].Name, "one")
	assert.Empty(t, p.Elem[0].Key)
	assert.Equal(t, p.Elem[1].Name, "two")
	assert.Equal(t, map[string]string{"a": "b"}, p.Elem[1].Key)
	assert.Equal(t, p.Elem[2].Name, "three")
	assert.Empty(t, p.Elem[2].Key)
}

func TestUpdateString(t *testing.T) {
	s := "stuff"
	u := UpdateString("/foo", "v1-device", &s)
	assert.NotNil(t, u)
	assert.NotNil(t, u.Path)
	assert.Equal(t, u.Path.Target, "v1-device")
	assert.Equal(t, u.Path.Elem[0].Name, "foo")
	assert.NotNil(t, u.Val)
	assert.Equal(t, "stuff", u.Val.GetStringVal())

	// nil value should return nil update
	u = UpdateString("/foo", "v1-device", nil)
	assert.Nil(t, u)
}

func TestUpdateUInt32(t *testing.T) {
	var i uint32 = 1234
	u := UpdateUInt32("/foo", "v1-device", &i)
	assert.NotNil(t, u)
	assert.NotNil(t, u.Path)
	assert.Equal(t, u.Path.Target, "v1-device")
	assert.Equal(t, u.Path.Elem[0].Name, "foo")
	assert.NotNil(t, u.Val)
	assert.Equal(t, uint64(1234), u.Val.GetUintVal())

	// nil value should return nil update
	u = UpdateUInt32("/foo", "v1-device", nil)
	assert.Nil(t, u)
}

func TestUpdateUInt64(t *testing.T) {
	var i uint64 = 1234
	u := UpdateUInt64("/foo", "v1-device", &i)
	assert.NotNil(t, u)
	assert.NotNil(t, u.Path)
	assert.Equal(t, u.Path.Target, "v1-device")
	assert.Equal(t, u.Path.Elem[0].Name, "foo")
	assert.NotNil(t, u.Val)
	assert.Equal(t, uint64(1234), u.Val.GetUintVal())

	// nil value should return nil update
	u = UpdateUInt64("/foo", "v1-device", nil)
	assert.Nil(t, u)
}

func TestUpdateBool(t *testing.T) {
	b := true
	u := UpdateBool("/foo", "v1-device", &b)
	assert.NotNil(t, u)
	assert.NotNil(t, u.Path)
	assert.Equal(t, u.Path.Target, "v1-device")
	assert.Equal(t, u.Path.Elem[0].Name, "foo")
	assert.NotNil(t, u.Val)
	assert.Equal(t, true, u.Val.GetBoolVal())

	// nil value should return nil update
	u = UpdateBool("/foo", "v1-device", nil)
	assert.Nil(t, u)
}

func TestAddUpdate(t *testing.T) {
	updates := []*gpb.Update{}

	// adding a nil update is a no-op
	updates = AddUpdate(updates, nil)
	assert.Len(t, updates, 0)

	// adding a non-nil update adds to the list
	b := true
	updates = AddUpdate(updates, UpdateBool("/foo", "v1-device", &b))
	assert.Len(t, updates, 1)
}

func TestDeletesFromUpdates(t *testing.T) {
	b := true
	updates := []*gpb.Update{}
	updates = AddUpdate(updates, UpdateBool("/foo", "v1-device", &b))
	updates = AddUpdate(updates, UpdateBool("/bar", "v1-device", &b))

	d := DeleteFromUpdates(updates, "v0-device")
	assert.NotNil(t, d)
	assert.Len(t, d, 2)
	assert.Equal(t, "v0-device", d[0].Target)
	assert.Equal(t, "foo", d[0].Elem[0].Name)
	assert.Equal(t, "v0-device", d[1].Target)
	assert.Equal(t, "bar", d[1].Elem[0].Name)
}
