// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0
package gnmi

import (
	"github.com/stretchr/testify/assert"
	"testing"

	pb "github.com/openconfig/gnmi/proto/gnmi"
)

func TestConvertTypedValueToJsonValue(t *testing.T) {
	// small unsigned integer may be returned as a uint64, json won't complain about it, as
	// the value itself fits in 32-bits.
	tv := &pb.TypedValue{Value: &pb.TypedValue_UintVal{UintVal: 123}}
	i, err := convertTypedValueToJsonValue(tv)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123), i.(uint64))

	// the largest uint that will fit in 32 bits
	tv = &pb.TypedValue{Value: &pb.TypedValue_UintVal{UintVal: 4294967295}}
	i, err = convertTypedValueToJsonValue(tv)
	assert.Nil(t, err)
	assert.Equal(t, uint64(4294967295), i.(uint64))

	// large integer should be returned as a string
	tv = &pb.TypedValue{Value: &pb.TypedValue_UintVal{UintVal: 4294967296}}
	i, err = convertTypedValueToJsonValue(tv)
	assert.Nil(t, err)
	assert.Equal(t, "4294967296", i.(string))

	// small unsigned integer may be returned as a int64, json won't complain about it, as
	// the value itself fits in 32-bits.
	tv = &pb.TypedValue{Value: &pb.TypedValue_IntVal{IntVal: 123}}
	i, err = convertTypedValueToJsonValue(tv)
	assert.Nil(t, err)
	assert.Equal(t, int64(123), i.(int64))

	// the largest uint that will fit in 32 bits
	tv = &pb.TypedValue{Value: &pb.TypedValue_IntVal{IntVal: 2147483647}}
	i, err = convertTypedValueToJsonValue(tv)
	assert.Nil(t, err)
	assert.Equal(t, int64(2147483647), i.(int64))

	// large integer should be returned as a string
	tv = &pb.TypedValue{Value: &pb.TypedValue_IntVal{IntVal: 2147483648}}
	i, err = convertTypedValueToJsonValue(tv)
	assert.Nil(t, err)
	assert.Equal(t, "2147483648", i.(string))
}
