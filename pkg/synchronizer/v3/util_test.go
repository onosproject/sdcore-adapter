// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package synchronizerv3

import (
	"github.com/stretchr/testify/assert"
	"testing"

	models_v3 "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
)

func TestFormatImsi(t *testing.T) {
	// straightforward conversion
	imsi, err := FormatImsi("CCCNNNEEESSSSSS", 123, 456, 789, 123456)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123456789123456), imsi)

	// zero padding on each field
	imsi, err = FormatImsi("CCCNNNEEESSSSSS", 12, 34, 56, 78)
	assert.Nil(t, err)
	assert.Equal(t, uint64(12034056000078), imsi)

	// forced zero after the MNC
	imsi, err = FormatImsi("CCCNN0EEESSSSSS", 123, 45, 789, 123456)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123450789123456), imsi)

	// subscriber is too long
	_, err = FormatImsi("CCCNNNEEESSSSSS", 123, 456, 789, 1234567)
	assert.EqualError(t, err, "Failed to convert all Subscriber digits")

	// unrecognized character
	_, err = FormatImsi("CCCNNNEEESSSSSZ", 123, 456, 789, 123456)
	assert.EqualError(t, err, "Unrecognized IMSI format specifier 'Z'")
}

func TestFormatImsiDef(t *testing.T) {
	i := &models_v3.Site_Site_Site_ImsiDefinition{
		Mcc:        aUint32(123),
		Mnc:        aUint32(45),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	imsi, err := FormatImsiDef(i, 123456)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123450789123456), imsi)

	// If format is nil, a default will be used
	i = &models_v3.Site_Site_Site_ImsiDefinition{
		Mcc:        aUint32(123),
		Mnc:        aUint32(45),
		Enterprise: aUint32(789),
		Format:     nil,
	}
	imsi, err = FormatImsiDef(i, 123456)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123045789123456), imsi)

	// Should reproduce the same errors as validateImsiDefinition

	// missing MCC
	i = &models_v3.Site_Site_Site_ImsiDefinition{
		Mnc:        aUint32(45),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	_, err = FormatImsiDef(i, 123456)
	assert.EqualError(t, err, "Format contains C, yet MCC is nil")

	// missing MNC
	i = &models_v3.Site_Site_Site_ImsiDefinition{
		Mcc:        aUint32(123),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	_, err = FormatImsiDef(i, 123456)
	assert.EqualError(t, err, "Format contains N, yet MNC is nil")

	// missing Ent
	i = &models_v3.Site_Site_Site_ImsiDefinition{
		Mcc:    aUint32(123),
		Mnc:    aUint32(45),
		Format: aStr("CCCNN0EEESSSSSS"),
	}
	_, err = FormatImsiDef(i, 123456)
	assert.EqualError(t, err, "Format contains E, yet Enterprise is nil")

	// Wrong number of characters
	i = &models_v3.Site_Site_Site_ImsiDefinition{
		Mcc:        aUint32(123),
		Mnc:        aUint32(45),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSS"),
	}
	_, err = FormatImsiDef(i, 123456)
	assert.EqualError(t, err, "Format is not 15 characters")
}

func TestProtoStringToProtoNumber(t *testing.T) {
	n, err := ProtoStringToProtoNumber("UDP")
	assert.Nil(t, err)
	assert.Equal(t, uint32(17), n)

	n, err = ProtoStringToProtoNumber("TCP")
	assert.Nil(t, err)
	assert.Equal(t, uint32(6), n)

	_, err = ProtoStringToProtoNumber("MQTT")
	assert.EqualError(t, err, "Unknown protocol MQTT")
}
