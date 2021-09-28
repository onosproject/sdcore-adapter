// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv3 implements declarations and utilities for the v3 synchronizer.
package synchronizerv4

import (
	"errors"
	"fmt"
	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
	"strings"
)

// Validation functions, return an error if the given struct is missing data that
// prevents synchronization.

// TODO: See if there is a way to do this automatically using ygot

// return error if VCS cannot be synchronized due to missing data
func validateVcs(vcs *models.Vcs_Vcs_Vcs) error {
	if vcs.Sst == nil {
		return fmt.Errorf("Sst is nil")
	}
	return nil
}

// return error if IpDomain cannot be synchronized due to missing data
func validateAppEndpoint(ep *models.Application_Application_Application_Endpoint) error {
	if ep.Address == nil {
		return fmt.Errorf("Address is nil")
	}
	if ep.PortStart == nil {
		return fmt.Errorf("PortStart is nil")
	}
	return nil
}

// return error if IpDomain cannot be synchronized due to missing data
func validateIPDomain(ipd *models.IpDomain_IpDomain_IpDomain) error {
	if ipd.Subnet == nil {
		return fmt.Errorf("Subnet is nil")
	}
	return nil
}

// return error if AccessPoint cannot be synchronized due to missing data
func validateAccessPoint(ap *models.ApList_ApList_ApList_AccessPoints) error {
	if ap.Address == nil {
		return fmt.Errorf("Address is nil")
	}
	if ap.Tac == nil {
		return fmt.Errorf("Tac is nil")
	}
	return nil
}

// return error if UPF cannot be synchronized due to missing data
func validateUpf(u *models.Upf_Upf_Upf) error {
	if u.Address == nil {
		return fmt.Errorf("Address is nil")
	}
	if u.Port == nil {
		return fmt.Errorf("Port is nil")
	}
	return nil
}

func validateImsiDefinition(i *models.Site_Site_Site_ImsiDefinition) error {
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
