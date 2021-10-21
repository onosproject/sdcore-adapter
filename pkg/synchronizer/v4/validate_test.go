// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv4 implements the version 4 synchronizer.
package synchronizerv4

import (
	"github.com/stretchr/testify/assert"
	"testing"

	models_v4 "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
)

func TestValidateVcs(t *testing.T) {
	v := &models_v4.OnfVcs_Vcs_Vcs{
		Sst:             aUint8(123),
		Sd:              aUint32(456),
		DefaultBehavior: aStr("DENY-ALL"),
	}
	err := validateVcs(v)
	assert.Nil(t, err)

	v = &models_v4.OnfVcs_Vcs_Vcs{
		Sd:              aUint32(456),
		DefaultBehavior: aStr("DENY-ALL"),
	}
	err = validateVcs(v)
	assert.EqualError(t, err, "Sst is nil")

	// SD is optional
	v = &models_v4.OnfVcs_Vcs_Vcs{
		Sst:             aUint8(123),
		DefaultBehavior: aStr("DENY-ALL"),
	}
	err = validateVcs(v)
	assert.Nil(t, err)
}

func TestValidateAppEndpoint(t *testing.T) {
	e := &models_v4.OnfApplication_Application_Application_Endpoint{
		PortStart: aUint16(123),
	}
	err := validateAppEndpoint(e)
	assert.Nil(t, err)

	e = &models_v4.OnfApplication_Application_Application_Endpoint{}
	err = validateAppEndpoint(e)
	assert.EqualError(t, err, "PortStart is nil")
}

func TestValidateIPDomain(t *testing.T) {
	i := &models_v4.OnfIpDomain_IpDomain_IpDomain{
		Subnet: aStr("1.2.3.4/24"),
	}
	err := validateIPDomain(i)
	assert.Nil(t, err)

	// Missing subnet
	i = &models_v4.OnfIpDomain_IpDomain_IpDomain{}
	err = validateIPDomain(i)
	assert.EqualError(t, err, "Subnet is nil")
}

func TestValidateSmallCell(t *testing.T) {
	a := &models_v4.OnfSite_Site_Site_SmallCell{
		Address: aStr("1.2.3.4"),
		Tac:     aStr("1234"),
	}
	err := validateSmallCell(a)
	assert.Nil(t, err)

	// missing address
	a = &models_v4.OnfSite_Site_Site_SmallCell{
		Tac: aStr("1234"),
	}
	err = validateSmallCell(a)
	assert.EqualError(t, err, "Address is nil")

	// missing Tac
	a = &models_v4.OnfSite_Site_Site_SmallCell{
		Address: aStr("1.2.3.4"),
	}
	err = validateSmallCell(a)
	assert.EqualError(t, err, "Tac is nil")
}

func TestValidateUpf(t *testing.T) {
	u := &models_v4.OnfUpf_Upf_Upf{
		Address: aStr("1.2.3.4"),
		Port:    aUint16(1234),
	}
	err := validateUpf(u)
	assert.Nil(t, err)

	// missing address
	u = &models_v4.OnfUpf_Upf_Upf{
		Port: aUint16(1234),
	}
	err = validateUpf(u)
	assert.EqualError(t, err, "Address is nil")

	// missing port
	u = &models_v4.OnfUpf_Upf_Upf{
		Address: aStr("1.2.3.4"),
	}
	err = validateUpf(u)
	assert.EqualError(t, err, "Port is nil")
}

func TestValidateImsiDefinition(t *testing.T) {
	i := &models_v4.OnfSite_Site_Site_ImsiDefinition{
		Mcc:        aStr("123"),
		Mnc:        aStr("45"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	err := validateImsiDefinition(i)
	assert.Nil(t, err)

	// missing MCC
	i = &models_v4.OnfSite_Site_Site_ImsiDefinition{
		Mnc:        aStr("45"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	err = validateImsiDefinition(i)
	assert.EqualError(t, err, "Format contains C, yet MCC is nil")

	// missing MNC
	i = &models_v4.OnfSite_Site_Site_ImsiDefinition{
		Mcc:        aStr("123"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSSS"),
	}
	err = validateImsiDefinition(i)
	assert.EqualError(t, err, "Format contains N, yet MNC is nil")

	// missing Ent
	i = &models_v4.OnfSite_Site_Site_ImsiDefinition{
		Mcc:    aStr("123"),
		Mnc:    aStr("45"),
		Format: aStr("CCCNN0EEESSSSSS"),
	}
	err = validateImsiDefinition(i)
	assert.EqualError(t, err, "Format contains E, yet Enterprise is nil")

	// Wrong number of characters
	i = &models_v4.OnfSite_Site_Site_ImsiDefinition{
		Mcc:        aStr("123"),
		Mnc:        aStr("45"),
		Enterprise: aUint32(789),
		Format:     aStr("CCCNN0EEESSSSS"),
	}
	err = validateImsiDefinition(i)
	assert.EqualError(t, err, "Format is not 15 characters")

	// Default format is okay
	i = &models_v4.OnfSite_Site_Site_ImsiDefinition{
		Mcc:        aStr("123"),
		Mnc:        aStr("45"),
		Enterprise: aUint32(789),
	}
	err = validateImsiDefinition(i)
	assert.Nil(t, err)
}
