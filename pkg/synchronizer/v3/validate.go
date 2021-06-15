// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizerv3

import (
	"fmt"
	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
)

// Validation functions, return an error if the given struct is missing data that
// prevents synchronization.

// TODO: See if there is a way to do this automatically using ygot

// return error if VCS cannot be synchronized due to missing data
func (s *Synchronizer) validateVcs(vcs *models.Vcs_Vcs_Vcs) error {
	if vcs.Sst == nil {
		return fmt.Errorf("Sst is nil")
	}
	if vcs.Sd == nil {
		return fmt.Errorf("Sd is nil")
	}
	return nil
}

// return error if IpDomain cannot be synchronized due to missing data
func (s *Synchronizer) validateAppEndpoint(ep *models.Application_Application_Application_Endpoint) error {
	if ep.Address == nil {
		return fmt.Errorf("Address is nil")
	}
	if ep.PortStart == nil {
		return fmt.Errorf("PortStart is nil")
	}
	return nil
}

// return error if IpDomain cannot be synchronized due to missing data
func (s *Synchronizer) validateIpDomain(ipd *models.IpDomain_IpDomain_IpDomain) error {
	if ipd.Subnet == nil {
		return fmt.Errorf("Subnet is nil")
	}
	return nil
}

// return error if AccessPoint cannot be synchronized due to missing data
func (s *Synchronizer) validateAccessPoint(ap *models.ApList_ApList_ApList_AccessPoints) error {
	if ap.Address == nil {
		return fmt.Errorf("Address is nil")
	}
	if ap.Tac == nil {
		return fmt.Errorf("Tac is nil")
	}
	return nil
}

// return error if Network cannot be synchronized due to missing data
func (s *Synchronizer) validateNetwork(n *models.Network_Network_Network) error {
	if n.Mnc == nil {
		return fmt.Errorf("Mnc is nil")
	}
	if n.Mcc == nil {
		return fmt.Errorf("Mcc is nil")
	}
	return nil
}

// return error if UPF cannot be synchronized due to missing data
func (s *Synchronizer) validateUpf(u *models.Upf_Upf_Upf) error {
	if u.Address == nil {
		return fmt.Errorf("Address is nil")
	}
	if u.Port == nil {
		return fmt.Errorf("Port is nil")
	}
	return nil
}
