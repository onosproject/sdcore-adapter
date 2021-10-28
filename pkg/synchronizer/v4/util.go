// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv4 implements declarations and utilities for the v4 synchronizer.
package synchronizerv4

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
	"github.com/onosproject/sdcore-adapter/pkg/synchronizer"
)

// FormatImsi formats MCC, MNC, ENT, and SUB into an IMSI, according to a format specifier
func FormatImsi(format string, mcc string, mnc string, ent uint32, sub uint64) (uint64, error) {
	var imsi uint64
	var mult uint64
	var mccUint uint64
	var mncUint uint64
	var err error
	mult = 1

	if mcc != "" {
		mccUint, err = strconv.ParseUint(mcc, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse mcc: %v", err)
		}
	} else {
		mccUint = 0
	}

	if mnc != "" {
		mncUint, err = strconv.ParseUint(mnc, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("failed to parse mnc: %v", err)
		}
	} else {
		mncUint = 0
	}

	// Build the IMSI from right to left, as it makes it easy to convert and pad integers
	for i := len(format) - 1; i >= 0; i-- {
		switch format[i] {
		case 'C':
			imsi = imsi + mccUint%10*mult
			mult *= 10
			mccUint = mccUint / 10
		case 'N':
			imsi = imsi + mncUint%10*mult
			mult *= 10
			mncUint = mncUint / 10
		case 'E':
			imsi = imsi + uint64(ent%10)*mult
			mult *= 10
			ent = ent / 10
		case 'S':
			imsi = imsi + (sub%10)*mult
			mult *= 10
			sub = sub / 10
		case '0':
			mult *= 10
		default:
			return 0, fmt.Errorf("Unrecognized IMSI format specifier '%c'", format[i])
		}
	}
	// IF there are any bits left in any of the fields, then it means we
	// had more digits than the IMSI format called for.
	if mccUint > 0 && strings.Contains(format, "C") {
		return 0, errors.New("Failed to convert all MCC digits")
	}
	if mncUint > 0 && strings.Contains(format, "N") {
		return 0, errors.New("Failed to convert all MNC digits")
	}
	if ent > 0 && strings.Contains(format, "E") {
		return 0, errors.New("Failed to convert all Enterprise digits")
	}
	if sub > 0 && strings.Contains(format, "S") {
		return 0, errors.New("Failed to convert all Subscriber digits")
	}

	return imsi, nil
}

// FormatImsiDef is a wrapper around FormatImsi that takes the ImsiDefinition gNMI instead of a set of arguments
func FormatImsiDef(i *models.OnfSite_Site_Site_ImsiDefinition, sub uint64) (uint64, error) {
	var format string
	if i.Format != nil {
		format = *i.Format
	} else {
		// default format from YANG
		format = DefaultImsiFormat
	}

	if err := validateImsiDefinition(i); err != nil {
		return 0, err
	}

	return FormatImsi(format,
		synchronizer.DerefStrPtr(i.Mcc, "0"),
		synchronizer.DerefStrPtr(i.Mnc, "0"),
		synchronizer.DerefUint32Ptr(i.Enterprise, 0),
		sub)
}

// MaskSubscriberImsi masks off any leading subscriber digits
func MaskSubscriberImsi(format string, sub uint64) (uint64, error) {
	var imsi uint64
	var mult uint64
	mult = 1
	// Build the IMSI from right to left, as it makes it easy to convert and pad integers
	for i := len(format) - 1; i >= 0; i-- {
		switch format[i] {
		case 'S':
			imsi = imsi + (sub%10)*mult
			mult *= 10
			sub = sub / 10
		default:
		}
	}
	return imsi, nil
}

// MaskSubscriberImsiDef is a wrapper around MaskSubscriberImsi that takes the ImsiDefinition gNMI instead of a set of arguments
func MaskSubscriberImsiDef(i *models.OnfSite_Site_Site_ImsiDefinition, sub uint64) (uint64, error) {
	var format string
	if i.Format != nil {
		format = *i.Format
	} else {
		// default format from YANG
		format = DefaultImsiFormat
	}

	if err := validateImsiDefinition(i); err != nil {
		return 0, err
	}

	return MaskSubscriberImsi(format, sub)
}

// ProtoStringToProtoNumber converts a protocol name to a number
func ProtoStringToProtoNumber(s string) (uint8, error) {
	n, okay := map[string]uint8{"TCP": 6, "UDP": 17}[s]
	if !okay {
		return 0, fmt.Errorf("Unknown protocol %s", s)
	}
	return n, nil
}

// aStr facilitates easy declaring of pointers to strings
func aStr(s string) *string {
	return &s
}

// aBool facilitates easy declaring of pointers to bools
func aBool(b bool) *bool {
	return &b
}

// aInt8 facilitates easy declaring of pointers to int8
func aInt8(u int8) *int8 {
	return &u
}

// aUint8 facilitates easy declaring of pointers to uint8
func aUint8(u uint8) *uint8 {
	return &u
}

// aUint16 facilitates easy declaring of pointers to uint16
func aUint16(u uint16) *uint16 {
	return &u
}

// aUint32 facilitates easy declaring of pointers to uint32
func aUint32(u uint32) *uint32 {
	return &u
}

// aUint64 facilitates easy declaring of pointers to uint64
func aUint64(u uint64) *uint64 {
	return &u
}
