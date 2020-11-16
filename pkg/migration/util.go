// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

/*
 * Various utility functions for migration.
 */

package migration

import (
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"strings"
)

// Given a string foo[k=v], return (foo, &k, &v)
// If the string does not contain a key, return (foo, nil, nil)
func SplitKey(name string) (*string, *string, *string) {
	parts := strings.Split(name, "[")
	name = parts[0]
	if name == "" {
		return nil, nil, nil
	}
	if len(parts) < 2 {
		return &name, nil, nil
	}

	keyValue := strings.TrimRight(parts[1], "]")
	parts = strings.Split(keyValue, "=")
	if len(parts) < 2 {
		return &name, nil, nil
	} else {
		key := parts[0]
		value := parts[1]
		return &name, &key, &value
	}
}

// convert a string for the format x/y/z[k=v] into a gdb.Path
func StringToPath(s string, target string) *gpb.Path {
	elems := []*gpb.PathElem{}

	parts := strings.Split(s, "/")

	for _, name := range parts {
		if len(name) > 0 {
			var keys map[string]string

			name, key, value := SplitKey(name)
			if name == nil {
				// the term was empty
				continue
			}

			if key != nil {
				keys = map[string]string{*key: *value}
			}

			elem := &gpb.PathElem{Name: *name,
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

// Given a list of Updates, create a corresponding list of deletes
func DeleteFromUpdates(updates []*gpb.Update, target string) []*gpb.Path {
	deletePaths := []*gpb.Path{}
	for _, update := range updates {
		deletePaths = append(deletePaths, &gpb.Path{
			Target: target,
			Elem:   update.Path.Elem,
		})
	}
	return deletePaths
}
