// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Utility functions for synchronizer
package synchronizerv3

import (
	"errors"
	"fmt"
	"strings"

	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
)

// Format MCC, MNC, ENT, and SUB into an IMSI, according to a format specifier
func FormatImsi(format string, mcc uint32, mnc uint32, ent uint32, sub uint64) (uint64, error) {
	var imsi uint64
	var mult uint64
	mult = 1
	// Build the IMSI from right to left, as it makes it easy to convert and pad integers
	for i := len(format) - 1; i >= 0; i-- {
		switch format[i] {
		case 'C':
			imsi = imsi + uint64(mcc%10)*mult
			mult *= 10
			mcc = mcc / 10
		case 'N':
			imsi = imsi + uint64(mnc%10)*mult
			mult *= 10
			mnc = mnc / 10
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
	if mcc > 0 && strings.Contains(format, "C") {
		return 0, errors.New("Failed to convert all MCC digits")
	}
	if mnc > 0 && strings.Contains(format, "N") {
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

// Wrapper around FormatImsi that takes the ImsiDefinition gNMI instead of a set of arguments
func FormatImsiDef(i *models.Site_Site_Site_ImsiDefinition, sub uint64) (uint64, error) {
	var format string
	if i.Format != nil {
		format = *i.Format
	} else {
		// default format from YANG
		format = DEFAULT_IMSI_FORMAT
	}

	if err := validateImsiDefinition(i); err != nil {
		return 0, err
	}

	//TODO for default site MCC,MNC,Ent Id values should be set to 0 instead they are returning as nil
	// following nil check is workaround
	if i.Mcc == nil {
		return FormatImsi(format, 0, 0, 0, sub)
	}

	return FormatImsi(format, *i.Mcc, *i.Mnc, *i.Enterprise, sub)
}

// Mask off any leading subscriber digits
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

// Wrapper around MaskSubscriberImsi that takes the ImsiDefinition gNMI instead of a set of arguments
func MaskSubscriberImsiDef(i *models.Site_Site_Site_ImsiDefinition, sub uint64) (uint64, error) {
	var format string
	if i.Format != nil {
		format = *i.Format
	} else {
		// default format from YANG
		format = DEFAULT_IMSI_FORMAT
	}

	if err := validateImsiDefinition(i); err != nil {
		return 0, err
	}

	return MaskSubscriberImsi(format, sub)
}

func ProtoStringToProtoNumber(s string) (uint32, error) {
	n, okay := map[string]uint32{"TCP": 6, "UDP": 17}[s]
	if !okay {
		return 0, fmt.Errorf("Unknown protocol %s", s)
	}
	return n, nil
}
