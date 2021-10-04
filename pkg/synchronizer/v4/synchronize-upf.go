// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv4 implements a synchronizer for converting sdcore gnmi to json
package synchronizerv4

import (
	"encoding/json"
	"fmt"

	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
)

type sliceQos struct {
	Uplink        uint64 `json:"uplinkMBR"`
	Downlink      uint64 `json:"downlinkMBR"`
	UplinkBurst   uint64 `json:"uplinkBurstSize"`
	DownlinkBurst uint64 `json:"downlinkBurstSize"`
}

type ueResourceInfo struct {
	Pool string `json:"uePoolId"`
	DNN  string `json:"dnn"`
}

type upfSliceConfig struct {
	SliceName      string           `json:"sliceName"`
	SliceQos       sliceQos         `json:"sliceQos"`
	UEResourceInfo []ueResourceInfo `json:"ueResourceInfo"`
}

// SynchronizeVcs synchronizes the VCSes to the UPF
// Return a count of push-related errors
func (s *Synchronizer) SynchronizeVcsUPF(device *models.Device, vcs *models.OnfVcs_Vcs_Vcs) (int, error) {
	if vcs.Upf == nil {
		return 0, fmt.Errorf("Vcs %s has no UPFs to synchronize", *vcs.Id)
	}

	aUpf, err := s.GetUpf(device, vcs.Upf)
	if err != nil {
		return 0, fmt.Errorf("Vcs %s unable to determine upf: %s", *vcs.Id, err)
	}
	err = validateUpf(aUpf)
	if err != nil {
		return 0, fmt.Errorf("Vcs %s Upf is invalid: %s", *vcs.Id, err)
	}

	if aUpf.ConfigEndpoint == nil {
		return 0, fmt.Errorf("Vcs %s UPF has no configuration endpoint", *vcs.Id)
	}

	sc := &upfSliceConfig{
		SliceName: *vcs.Id,
	}

	if (vcs.Slice != nil) && (vcs.Slice.Mbr != nil) {
		if vcs.Slice.Mbr.Uplink != nil {
			sc.SliceQos.Uplink = *vcs.Slice.Mbr.Uplink
		}
		if vcs.Slice.Mbr.Downlink != nil {
			sc.SliceQos.Uplink = *vcs.Slice.Mbr.Uplink
		}
	}

	dgList, err := s.GetVcsDG(device, vcs)
	if err != nil {
		return 0, fmt.Errorf("Vcs %s unable to determine dgList: %s", *vcs.Id, err)
	}

	for _, dg := range dgList {
		ipd, err := s.GetIPDomain(device, dg.IpDomain)
		if err != nil {
			return 0, fmt.Errorf("DeviceGroup %s failed to get IpDomain: %s", *dg.Id, err)
		}

		ueRes := ueResourceInfo{Pool: *dg.Id,
			DNN: *ipd.Dnn}
		sc.UEResourceInfo = append(sc.UEResourceInfo, ueRes)
	}

	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("Vcs %s failed to marshal UPF JSON: %s", *vcs.Id, err)
	}

	url := fmt.Sprintf("%s/v1/network-slice/%s", *aUpf.ConfigEndpoint, *vcs.Id)
	err = s.pusher.PushUpdate(url, data)
	if err != nil {
		return 1, fmt.Errorf("vcs %s failed to push UPF JSON: %s", *vcs.Id, err)
	}

	return 0, nil
}
