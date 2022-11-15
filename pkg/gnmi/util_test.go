// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
package gnmi

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"

	testmodels "github.com/onosproject/config-models/models/testdevice-2.0.x/api"
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

func TestGetChildNode(t *testing.T) {
	data := map[string]interface{}{}
	schema, err := testmodels.UnzipSchema()
	assert.Nil(t, err)

	// Check for nonexistent path
	elem := &pb.PathElem{Name: "doest_not_exist"}
	nextNode, nextSchema := getChildNode(data, schema["Device"], elem, true)
	assert.Nil(t, nextNode)
	assert.Nil(t, nextSchema)

	// outer container
	elem = &pb.PathElem{Name: "cont1a"}
	cont1aData, cont1aSchema := getChildNode(data, schema["Device"], elem, true)
	assert.NotNil(t, cont1aData)
	assert.NotNil(t, cont1aSchema)
	cont1aNode, ok := cont1aData.(map[string]interface{})
	assert.True(t, ok)

	// inner container
	elem = &pb.PathElem{Name: "cont2d"}
	cont2dData, cont2dSchema := getChildNode(cont1aNode, cont1aSchema, elem, true)
	assert.NotNil(t, cont2dData)
	assert.NotNil(t, cont2dSchema)
	cont2dNode, ok := cont2dData.(map[string]interface{})
	assert.True(t, ok)

	// choice
	elem = &pb.PathElem{Name: "pretzel"}
	pretzelData, pretzelSchema := getChildNode(cont2dNode, cont2dSchema, elem, true)
	assert.NotNil(t, pretzelData)
	assert.NotNil(t, pretzelSchema)

	// verify that the generate JSON is correct
	jsonStr, err := json.Marshal(data)
	assert.Nil(t, err)
	assert.JSONEq(t, `{"cont1a": {"cont2d": {"pretzel": {}}}}`, string(jsonStr))
}
