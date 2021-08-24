// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0
package gnmi

import (
	"github.com/stretchr/testify/assert"
	"testing"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

func TestConvertTypedValueToJSONValue(t *testing.T) {
	// Note: This function has been simplified and now relies on the caller
	// to specify whether or not the value needs to be converted to a
	// string.

	tv := &pb.TypedValue{Value: &pb.TypedValue_UintVal{UintVal: 123}}
	i, err := convertTypedValueToJSONValue(tv, false)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123), i.(uint64))

	tv = &pb.TypedValue{Value: &pb.TypedValue_UintVal{UintVal: 123}}
	i, err = convertTypedValueToJSONValue(tv, true)
	assert.Nil(t, err)
	assert.Equal(t, "123", i.(string))
}
