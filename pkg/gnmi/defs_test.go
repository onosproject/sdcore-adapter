// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0
package gnmi

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// Ensure that TestConfigCallbackType.String() does not fail for all valid enum values
func TestConfigCallbackType(t *testing.T) {
	for ct := ConfigCallbackType(0); ct < configCallbackTypeLimit; ct++ {
		assert.NotEmpty(t, ct.String())
	}
}
