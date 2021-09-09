// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package synchronizer

// BoolToUint32 convert a boolean to an unsigned integer
func BoolToUint32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

// DerefStrPtr dereference a string pointer, returning a default if it is nil
func DerefStrPtr(s *string, def string) string {
	if s == nil {
		return def
	}
	return *s
}

// DerefUint32Ptr dereference a uint32 pointer, returning default if it is nil
func DerefUint32Ptr(u *uint32, def uint32) uint32 {
	if u == nil {
		return def
	}
	return *u
}

// DerefUint16Ptr dereference a uint16 pointer, returning default if it is nil
func DerefUint16Ptr(u *uint16, def uint16) uint16 {
	if u == nil {
		return def
	}
	return *u
}
