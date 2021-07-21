// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package synchronizerv3

import (
	"github.com/stretchr/testify/assert"
	"testing"

	models_v3 "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
)

func TestValidateVcs(t *testing.T) {
	v := &models_v3.Vcs_Vcs_Vcs{
		Sst: aUint32(123),
		Sd:  aUint32(456),
	}
	err := validateVcs(v)
	assert.Nil(t, err)

	v = &models_v3.Vcs_Vcs_Vcs{
		Sd: aUint32(456),
	}
	err = validateVcs(v)
	assert.EqualError(t, err, "Sst is nil")

	v = &models_v3.Vcs_Vcs_Vcs{
		Sst: aUint32(123),
	}
	err = validateVcs(v)
	assert.EqualError(t, err, "Sd is nil")
}

func TestValidateAppEndpoint(t *testing.T) {
	e := &models_v3.Application_Application_Application_Endpoint{
		Address:   aStr("1.2.3.4"),
		PortStart: aUint32(123),
	}
	err := validateAppEndpoint(e)
	assert.Nil(t, err)

	e = &models_v3.Application_Application_Application_Endpoint{
		PortStart: aUint32(123),
	}
	err = validateAppEndpoint(e)
	assert.EqualError(t, err, "Address is nil")

	e = &models_v3.Application_Application_Application_Endpoint{
		Address: aStr("1.2.3.4"),
	}
	err = validateAppEndpoint(e)
	assert.EqualError(t, err, "PortStart is nil")
}

func TestValidateIpDomain(t *testing.T) {
	i := &models_v3.IpDomain_IpDomain_IpDomain{
		Subnet: aStr("1.2.3.4/24"),
	}
	err := validateIpDomain(i)
	assert.Nil(t, err)

	// Missing subnet
	i = &models_v3.IpDomain_IpDomain_IpDomain{}
	err = validateIpDomain(i)
	assert.EqualError(t, err, "Subnet is nil")
}

func TestValidateAccessPoint(t *testing.T) {
	a := &models_v3.ApList_ApList_ApList_AccessPoints{
		Address: aStr("1.2.3.4"),
		Tac:     aUint32(1234),
	}
	err := validateAccessPoint(a)
	assert.Nil(t, err)

	// missing address
	a = &models_v3.ApList_ApList_ApList_AccessPoints{
		Tac: aUint32(1234),
	}
	err = validateAccessPoint(a)
	assert.EqualError(t, err, "Address is nil")

	// missing Tac
	a = &models_v3.ApList_ApList_ApList_AccessPoints{
		Address: aStr("1.2.3.4"),
	}
	err = validateAccessPoint(a)
	assert.EqualError(t, err, "Tac is nil")
}

func TestValidateUpf(t *testing.T) {
	u := &models_v3.Upf_Upf_Upf{
		Address: aStr("1.2.3.4"),
		Port:    aUint32(1234),
	}
	err := validateUpf(u)
	assert.Nil(t, err)

	// missing address
	u = &models_v3.Upf_Upf_Upf{
		Port: aUint32(1234),
	}
	err = validateUpf(u)
	assert.EqualError(t, err, "Address is nil")

	// missing port
	u = &models_v3.Upf_Upf_Upf{
		Address: aStr("1.2.3.4"),
	}
	err = validateUpf(u)
	assert.EqualError(t, err, "Port is nil")
}

func TestValidateImsiDefinition(t *testing.T) {
	i := &models_v3.Site_Site_Site_ImsiDefinition{
		Mcc:        aUint32(123),
		Mnc:        aUint32(45),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	err := validateImsiDefinition(i)
	assert.Nil(t, err)

	// missing MCC
	i = &models_v3.Site_Site_Site_ImsiDefinition{
		Mnc:        aUint32(45),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	err = validateImsiDefinition(i)
	assert.EqualError(t, err, "Format contains C, yet MCC is nil")

	// missing MNC
	i = &models_v3.Site_Site_Site_ImsiDefinition{
		Mcc:        aUint32(123),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	err = validateImsiDefinition(i)
	assert.EqualError(t, err, "Format contains N, yet MNC is nil")

	// missing Ent
	i = &models_v3.Site_Site_Site_ImsiDefinition{
		Mcc:    aUint32(123),
		Mnc:    aUint32(45),
		Format: aStr("CCCNN0EEESSSSSS"),
	}
	err = validateImsiDefinition(i)
	assert.EqualError(t, err, "Format contains E, yet Enterprise is nil")

	// Wrong number of characters
	i = &models_v3.Site_Site_Site_ImsiDefinition{
		Mcc:        aUint32(123),
		Mnc:        aUint32(45),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSS"),
	}
	err = validateImsiDefinition(i)
	assert.EqualError(t, err, "Format is not 15 characters")

	// Default format is okay
	i = &models_v3.Site_Site_Site_ImsiDefinition{
		Mcc:        aUint32(123),
		Mnc:        aUint32(45),
		Enterprise: aUint32(789),
	}
	err = validateImsiDefinition(i)
	assert.Nil(t, err)
}
