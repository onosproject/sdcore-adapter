// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package synchronizer

func BoolToUint32(b bool) uint32 {
	if b {
		return 1
	} else {
		return 0
	}
}

func DerefStrPtr(s *string, def string) string {
	if s == nil {
		return def
	}
	return *s
}

func DerefUint32Ptr(u *uint32, def uint32) uint32 {
	if u == nil {
		return def
	}
	return *u
}
