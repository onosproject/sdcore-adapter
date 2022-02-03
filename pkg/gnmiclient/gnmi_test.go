// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package gnmiclient

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewGnmi(t *testing.T) {
	gnmi, err := NewGnmi("onos-config:5150", time.Second*15)
	assert.Error(t, err)
	assert.Equal(t, "could not create a gNMI client: Dialer(onos-config:5150, 15s): context deadline exceeded", err.Error())
	assert.Nil(t, gnmi)
}

func TestNewGnmiWithInterceptor(t *testing.T) {
	gnmi, token, err := NewGnmiWithInterceptor("onos-config:5150", time.Second*15)
	assert.NoError(t, err)
	assert.Equal(t, "onos-config:5150", gnmi.Address())
	assert.Equal(t, "", token)
}
