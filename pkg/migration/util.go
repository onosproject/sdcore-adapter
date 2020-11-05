// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package migration

import (
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"strings"
)

func SplitKey(name string) (string, string, string) {
	parts := strings.Split(name, "[")
	name = parts[0]
	keyValue := strings.TrimRight(parts[1], "]")
	parts = strings.Split(keyValue, "=")
	key := parts[0]
	value := parts[1]
	return name, key, value
}

// convert a string for the format x/y/z[k=v] into a gdb.Path
func StringToPath(s string, target string) *gpb.Path {
	elems := []*gpb.PathElem{}

	parts := strings.Split(s, "/")

	for _, name := range parts {
		if len(name) > 0 {
			var keys map[string]string

			// see if there is a key in the format [x=y] and if so,
			// parse and remove it.
			if strings.Contains(name, "[") {
				splitName, key, value := SplitKey(name)
				name = splitName
				keys = map[string]string{key: value}
			}

			elem := &gpb.PathElem{Name: name,
				Key: keys}

			elems = append(elems, elem)
		}
	}

	return &gpb.Path{
		Target: target,
		Elem:   elems,
	}
}

// Create a gpb.Update for a string value
func UpdateString(path string, target string, val *string) *gpb.Update {
	if val == nil {
		return nil
	}

	return &gpb.Update{
		Path: StringToPath(path, target),
		Val:  &gpb.TypedValue{Value: &gpb.TypedValue_StringVal{StringVal: *val}},
	}
}

// Create a gpb.Update for a uint32 value
func UpdateUInt32(path string, target string, val *uint32) *gpb.Update {
	if val == nil {
		return nil
	}

	return &gpb.Update{
		Path: StringToPath(path, target),
		Val:  &gpb.TypedValue{Value: &gpb.TypedValue_UintVal{UintVal: uint64(*val)}},
	}
}

// Create a gpb.Update for a uint64 value
func UpdateUInt64(path string, target string, val *uint64) *gpb.Update {
	if val == nil {
		return nil
	}

	return &gpb.Update{
		Path: StringToPath(path, target),
		Val:  &gpb.TypedValue{Value: &gpb.TypedValue_UintVal{UintVal: *val}},
	}
}

// Create a gpb.Update for a bool value
func UpdateBool(path string, target string, val *bool) *gpb.Update {
	if val == nil {
		return nil
	}

	return &gpb.Update{
		Path: StringToPath(path, target),
		Val:  &gpb.TypedValue{Value: &gpb.TypedValue_BoolVal{BoolVal: *val}},
	}
}

// Add a gpb.Update to a list of updates, only if the gpb.Update is not
// nil.
func AddUpdate(updates []*gpb.Update, update *gpb.Update) []*gpb.Update {
	if update != nil {
		updates = append(updates, update)
	}
	return updates
}
