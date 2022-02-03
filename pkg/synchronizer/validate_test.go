// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements the synchronizer.
package synchronizer

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateSlice(t *testing.T) {
	v := &Slice{
		Sst:             aUint8(123),
		Sd:              aUint32(456),
		DefaultBehavior: aStr("DENY-ALL"),
	}
	err := validateSlice(v)
	assert.Nil(t, err)

	v = &Slice{
		Sd:              aUint32(456),
		DefaultBehavior: aStr("DENY-ALL"),
	}
	err = validateSlice(v)
	assert.EqualError(t, err, "Sst is nil")

	// SD is optional
	v = &Slice{
		Sst:             aUint8(123),
		DefaultBehavior: aStr("DENY-ALL"),
	}
	err = validateSlice(v)
	assert.Nil(t, err)
}

func TestValidateAppEndpoint(t *testing.T) {
	e := &ApplicationEndpoint{
		PortStart: aUint16(123),
	}
	err := validateAppEndpoint(e)
	assert.Nil(t, err)

	e = &ApplicationEndpoint{}
	err = validateAppEndpoint(e)
	assert.EqualError(t, err, "PortStart is nil")
}

func TestValidateIPDomain(t *testing.T) {
	i := &IpDomain{
		Subnet: aStr("1.2.3.4/24"),
	}
	err := validateIPDomain(i)
	assert.Nil(t, err)

	// Missing subnet
	i = &IpDomain{}
	err = validateIPDomain(i)
	assert.EqualError(t, err, "Subnet is nil")
}

func TestValidateSmallCell(t *testing.T) {
	a := &SmallCell{
		Address: aStr("1.2.3.4"),
		Tac:     aStr("1234"),
	}
	err := validateSmallCell(a)
	assert.Nil(t, err)

	// missing address
	a = &SmallCell{
		Tac: aStr("1234"),
	}
	err = validateSmallCell(a)
	assert.EqualError(t, err, "Address is nil")

	// missing Tac
	a = &SmallCell{
		Address: aStr("1.2.3.4"),
	}
	err = validateSmallCell(a)
	assert.EqualError(t, err, "Tac is nil")
}

func TestValidateUpf(t *testing.T) {
	u := &Upf{
		Address: aStr("1.2.3.4"),
		Port:    aUint16(1234),
	}
	err := validateUpf(u)
	assert.Nil(t, err)

	// missing address
	u = &Upf{
		Port: aUint16(1234),
	}
	err = validateUpf(u)
	assert.EqualError(t, err, "Address is nil")

	// missing port
	u = &Upf{
		Address: aStr("1.2.3.4"),
	}
	err = validateUpf(u)
	assert.EqualError(t, err, "Port is nil")
}

func TestValidateImsiDefinition(t *testing.T) {
	i := &ImsiDefinition{
		Mcc:        aStr("123"),
		Mnc:        aStr("45"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	err := validateImsiDefinition(i)
	assert.Nil(t, err)

	// missing MCC
	i = &ImsiDefinition{
		Mnc:        aStr("45"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	err = validateImsiDefinition(i)
	assert.EqualError(t, err, "Format contains C, yet MCC is nil")

	// missing MNC
	i = &ImsiDefinition{
		Mcc:        aStr("123"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	err = validateImsiDefinition(i)
	assert.EqualError(t, err, "Format contains N, yet MNC is nil")

	// missing Ent
	i = &ImsiDefinition{
		Mcc:    aStr("123"),
		Mnc:    aStr("45"),
		Format: aStr("CCCNN0EEESSSSSS"),
	}
	err = validateImsiDefinition(i)
	assert.EqualError(t, err, "Format contains E, yet Enterprise is nil")

	// Wrong number of characters
	i = &ImsiDefinition{
		Mcc:        aStr("123"),
		Mnc:        aStr("45"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSS"),
	}
	err = validateImsiDefinition(i)
	assert.EqualError(t, err, "Format is not 15 characters")

	// Default format is okay
	i = &ImsiDefinition{
		Mcc:        aStr("123"),
		Mnc:        aStr("45"),
		Enterprise: aUint32(789),
	}
	err = validateImsiDefinition(i)
	assert.Nil(t, err)
}
