// SPDX-FileCopyrightText: 2022-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package steps

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_isValidIdentifier(t *testing.T) {
	str1 := "1-abcd"
	matched := isValidIdentifier(str1)
	assert.True(t, !matched)
	str1 = "ascd-12-1"
	matched = isValidIdentifier(str1)
	assert.True(t, matched)
	str1 = "Ascd-12-1"
	matched = isValidIdentifier(str1)
	assert.True(t, !matched)
	str1 = "ascd-12-"
	matched = isValidIdentifier(str1)
	assert.True(t, !matched)
	str1 = "-scd-12-1"
	matched = isValidIdentifier(str1)
	assert.True(t, !matched)
	str1 = "_scd-12"
	matched = isValidIdentifier(str1)
	assert.True(t, !matched)
	str1 = "scd-bg12om-axz"
	matched = isValidIdentifier(str1)
	assert.True(t, matched)
}

func Test_convertIdentifier(t *testing.T) {
	str := "connectivity-service-v2"
	s := convertIdentifier(str)
	assert.Equal(t, "connectivity-service-v2", s)

	str = "ConnectivityService-V2"
	s = convertIdentifier(str)
	assert.Equal(t, "connectivity-service-v2", s)

	str = "BengaluruDroneCamera-2StatusLive"
	s = convertIdentifier(str)
	assert.Equal(t, "bengaluru-drone-camera-2-status-live", s)

	str = "BengaluruCamera-#2"
	s = convertIdentifier(str)
	assert.Equal(t, "bengaluru-camera-2", s)

	str = "123BengaluruCamera-#2"
	s = convertIdentifier(str)
	assert.Equal(t, "bengaluru-camera-2", s)

	str = "123bengaluruCamera-#2"
	s = convertIdentifier(str)
	assert.Equal(t, "bengaluru-camera-2", s)

	str = "-123#bengaluruCamera-#2"
	s = convertIdentifier(str)
	assert.Equal(t, "bengaluru-camera-2", s)
}
