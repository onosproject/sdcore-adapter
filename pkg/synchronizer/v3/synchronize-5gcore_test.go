// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package synchronizerv3

import (
	models_v3 "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

// to facilitate easy declaring of pointers to strings
func aStr(s string) *string {
	return &s
}

// to facilitate easy declaring of pointers to bools
func aBool(b bool) *bool {
	return &b
}

// to facilitate easy declaring of pointers to uint32
func aUint32(u uint32) *uint32 {
	return &u
}

// to facilitate easy declaring of pointers to uint64
func aUint64(u uint64) *uint64 {
	return &u
}

// populate an Enterprise structure
func MakeEnterprise(desc string, displayName string, id string, cs []string) *models_v3.Enterprise_Enterprise_Enterprise {
	csList := map[string]*models_v3.Enterprise_Enterprise_Enterprise_ConnectivityService{}

	for _, csId := range cs {
		csList[csId] = &models_v3.Enterprise_Enterprise_Enterprise_ConnectivityService{
			ConnectivityService: aStr(csId),
			Enabled:             aBool(true),
		}
	}

	ent := models_v3.Enterprise_Enterprise_Enterprise{
		Description:         aStr(desc),
		DisplayName:         aStr(displayName),
		Id:                  aStr(id),
		ConnectivityService: csList,
	}

	return &ent
}

func MakeCs(desc string, displayName string, id string) *models_v3.ConnectivityService_ConnectivityService_ConnectivityService {
	cs := models_v3.ConnectivityService_ConnectivityService_ConnectivityService{
		Description: aStr(desc),
		DisplayName: aStr(displayName),
		Id:          aStr(id),
	}

	return &cs
}

// an empty device should yield empty json
func TestSynchronizeDeviceEmpty(t *testing.T) {
	// Get a temporary file name and defer deletion of the file
	f, err := ioutil.TempFile("", "synchronizer-json")
	assert.Nil(t, err)
	tempFileName := f.Name()
	defer func() {
		os.Remove(tempFileName)
	}()

	s := Synchronizer{}
	s.SetOutputFileName(tempFileName)
	device := models_v3.Device{}
	err = s.SynchronizeDevice(&device)
	assert.Nil(t, err)

	content, err := ioutil.ReadFile(tempFileName)
	assert.Nil(t, err)
	assert.Equal(t, "", string(content))
}
