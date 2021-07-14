// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package synchronizerv3

import (
	"github.com/stretchr/testify/assert"
	"testing"
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
