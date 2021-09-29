// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv4 implements the version 4 synchronizer.
package synchronizerv4

import (
	"github.com/stretchr/testify/assert"
	"testing"

	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
)

func TestFormatImsi(t *testing.T) {
	// straightforward conversion
	imsi, err := FormatImsi("CCCNNNEEESSSSSS", "123", "456", 789, 123456)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123456789123456), imsi)

	// zero padding on each field
	imsi, err = FormatImsi("CCCNNNEEESSSSSS", "12", "34", 56, 78)
	assert.Nil(t, err)
	assert.Equal(t, uint64(12034056000078), imsi)

	// forced zero after the MNC
	imsi, err = FormatImsi("CCCNN0EEESSSSSS", "123", "45", 789, 123456)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123450789123456), imsi)

	// subscriber is too long
	_, err = FormatImsi("CCCNNNEEESSSSSS", "123", "456", 789, 1234567)
	assert.EqualError(t, err, "Failed to convert all Subscriber digits")

	// unrecognized character
	_, err = FormatImsi("CCCNNNEEESSSSSZ", "123", "456", 789, 123456)
	assert.EqualError(t, err, "Unrecognized IMSI format specifier 'Z'")
}

func TestFormatImsiDef(t *testing.T) {
	i := &models.OnfSite_Site_Site_ImsiDefinition{
		Mcc:        aStr("123"),
		Mnc:        aStr("45"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	imsi, err := FormatImsiDef(i, 123456)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123450789123456), imsi)

	// If format is nil, a default will be used
	i = &models.OnfSite_Site_Site_ImsiDefinition{
		Mcc:        aStr("123"),
		Mnc:        aStr("45"),
		Enterprise: aUint32(789),
		Format:     nil,
	}
	imsi, err = FormatImsiDef(i, 123456)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123045789123456), imsi)

	// Should reproduce the same errors as validateImsiDefinition

	// missing MCC
	i = &models.OnfSite_Site_Site_ImsiDefinition{
		Mnc:        aStr("45"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	_, err = FormatImsiDef(i, 123456)
	assert.EqualError(t, err, "Format contains C, yet MCC is nil")

	// missing MNC
	i = &models.OnfSite_Site_Site_ImsiDefinition{
		Mcc:        aStr("123"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	_, err = FormatImsiDef(i, 123456)
	assert.EqualError(t, err, "Format contains N, yet MNC is nil")

	// missing Ent
	i = &models.OnfSite_Site_Site_ImsiDefinition{
		Mcc:    aStr("123"),
		Mnc:    aStr("45"),
		Format: aStr("CCCNN0EEESSSSSS"),
	}
	_, err = FormatImsiDef(i, 123456)
	assert.EqualError(t, err, "Format contains E, yet Enterprise is nil")

	// Wrong number of characters
	i = &models.OnfSite_Site_Site_ImsiDefinition{
		Mcc:        aStr("123"),
		Mnc:        aStr("45"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSS"),
	}
	_, err = FormatImsiDef(i, 123456)
	assert.EqualError(t, err, "Format is not 15 characters")

	// 15-digit IMSI is just fine
	i = &models.OnfSite_Site_Site_ImsiDefinition{
		Mcc:        aStr("321"),
		Mnc:        aStr("54"),
		Enterprise: aUint32(987),
		Format:     aStr("SSSSSSSSSSSSSSS"),
	}
	imsi, err = FormatImsiDef(i, 123456789012345)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123456789012345), imsi)

	// Test bugfix on nil Enterprise
	i = &models.OnfSite_Site_Site_ImsiDefinition{
		Mcc:        aStr("321"),
		Mnc:        aStr("54"),
		Enterprise: nil,
		Format:     aStr("SSSSSSSSSSSSSSS"),
	}
	imsi, err = FormatImsiDef(i, 123456789012345)
	assert.Nil(t, err)
	assert.Equal(t, uint64(123456789012345), imsi)
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
