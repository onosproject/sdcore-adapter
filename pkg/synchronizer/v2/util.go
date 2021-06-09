// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Utility functions for synchronizer
package synchronizerv2

func boolToUint32(b bool) uint32 {
	if b {
		return 1
	} else {
		return 0
	}
}
