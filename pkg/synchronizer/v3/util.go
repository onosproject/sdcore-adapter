// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv3 implements declarations and utilities for the v3 synchronizer.
package synchronizerv3

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
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
func FormatImsiDef(i *models.Site_Site_Site_ImsiDefinition, sub uint64) (uint64, error) {
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
func MaskSubscriberImsiDef(i *models.Site_Site_Site_ImsiDefinition, sub uint64) (uint64, error) {
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
func ProtoStringToProtoNumber(s string) (uint32, error) {
	n, okay := map[string]uint32{"TCP": 6, "UDP": 17}[s]
	if !okay {
		return 0, fmt.Errorf("Unknown protocol %s", s)
	}
	return n, nil
}
