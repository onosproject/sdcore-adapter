// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizerv3

import (
	//"bytes"
	//"encoding/json"
	//"errors"
	//"fmt"
	"github.com/openconfig/ygot/ygot"
	//"net/http"
	//"os"
	//"time"

	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
)

/*
 * SPGW JSON Schema
 *
 * Note that while this is very close to the ROC modeling, there are slight
 * differences, and in general there may be translation between the ROC and
 * services managed by the ROC. Thus, we'll intentionally leave these
 * differences in place rather than try to force the ROC models and the SPGW
 * models to be identical.
 */

type IpDomain struct {
	Pool        string `json:"pool"`
	AdminStatus string `json:"admin-status"`
	DnsPrimary  string `json:"dns-primary"`
	Mtu         uint32 `json:"mtu"`
}

type DeviceGroup struct {
	Imsis        []string `json:"imsis"`
	IpDomainName string   `json:"ip-domain-name"`
	SiteInfo     string   `json:"site-info"`
	IpDomain     IpDomain `json:"ip-domain"`
}

func (s *Synchronizer) SynchronizeDevice(config ygot.ValidatedGoStruct) error {
	device := config.(*models.Device)

	_ = device
	return nil
}
