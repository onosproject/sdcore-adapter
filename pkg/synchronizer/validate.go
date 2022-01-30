// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements declarations and utilities for the synchronizer.
package synchronizer

import (
	"errors"
	"fmt"
	"strings"
)

// Validation functions, return an error if the given struct is missing data that
// prevents synchronization.

// TODO: See if there is a way to do this automatically using ygot

// return error if VCS cannot be synchronized due to missing data
func validateSlice(slice *Slice) error {
	if slice.Sst == nil {
		return fmt.Errorf("Sst is nil")
	}
	if slice.DefaultBehavior == nil {
		return fmt.Errorf("DefaultBehavior is nil")
	}
	return nil
}

// return error if IpDomain cannot be synchronized due to missing data
func validateAppEndpoint(ep *ApplicationEndpoint) error {
	if ep.PortStart == nil {
		return fmt.Errorf("PortStart is nil")
	}
	return nil
}

// return error if IpDomain cannot be synchronized due to missing data
func validateIPDomain(ipd *IpDomain) error {
	if ipd.Subnet == nil {
		return fmt.Errorf("Subnet is nil")
	}
	return nil
}

// return error if SmallCell cannot be synchronized due to missing data
func validateSmallCell(ap *SmallCell) error {
	if ap.Address == nil {
		return fmt.Errorf("Address is nil")
	}
	if ap.Tac == nil {
		return fmt.Errorf("Tac is nil")
	}
	return nil
}

// return error if UPF cannot be synchronized due to missing data
func validateUpf(u *Upf) error {
	if u.Address == nil {
		return fmt.Errorf("Address is nil")
	}
	if u.Port == nil {
		return fmt.Errorf("Port is nil")
	}
	return nil
}

func validateImsiDefinition(i *ImsiDefinition) error {
	var format string
	if i.Format != nil {
		format = *i.Format
	} else {
		// default format from YANG
		format = DefaultImsiFormat
	}

	if (i.Mcc == nil) && (strings.Contains(format, "C")) {
		return errors.New("Format contains C, yet MCC is nil")
	}
	if (i.Mnc == nil) && (strings.Contains(format, "N")) {
		return errors.New("Format contains N, yet MNC is nil")
	}
	if (i.Enterprise == nil) && (strings.Contains(format, "E")) {
		return errors.New("Format contains E, yet Enterprise is nil")
	}

	// Note: If format is nil, we'll assume it is the default
	if len(format) != 15 {
		return fmt.Errorf("Format is not 15 characters")
	}

	return nil
}

func validateDeviceGroup(dg *DeviceGroup) error {
	if dg.Device == nil {
		return fmt.Errorf("has no per-Device settings")
	}

	if dg.Mbr == nil {
		return fmt.Errorf("has per-Device settings, but no MBR")
	}

	if dg.Mbr.Uplink == nil {
		return fmt.Errorf("Device.MBR.Uplink is unset")
	}

	if dg.Mbr.Downlink == nil {
		return fmt.Errorf("Device.MBR.Downlink is unset")
	}

	if dg.Mbr.TrafficClass == nil {
		return fmt.Errorf("has no Device.Traffic-Class")
	}

	return nil
}
