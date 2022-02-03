// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

import (
	"encoding/json"
	"fmt"
)

type sliceQos struct {
	Uplink        uint64  `json:"uplinkMBR,omitempty"`
	Downlink      uint64  `json:"downlinkMBR,omitempty"`
	UplinkBurst   uint32  `json:"uplinkBurstSize,omitempty"`
	DownlinkBurst uint32  `json:"downlinkBurstSize,omitempty"`
	Unit          *string `json:"bitrateUnit,omitempty"`
}

type ueResourceInfo struct {
	Pool string `json:"uePoolId"`
	DNN  string `json:"dnn"`
}

type upfSliceConfig struct {
	SliceName      string           `json:"sliceName"`
	SliceQos       sliceQos         `json:"sliceQos"`
	UEResourceInfo []ueResourceInfo `json:"ueResourceInfo,omitempty"`
}

// SynchronizeSliceUPF synchronizes the VCSes to the UPF
// Return a count of push-related errors
func (s *Synchronizer) SynchronizeSliceUPF(scope *AetherScope, slice *Slice) (int, error) {
	if slice.Upf == nil {
		return 0, fmt.Errorf("Slice %s has no UPFs to synchronize", *slice.SliceId)
	}

	aUpf, err := s.GetUpf(scope, slice.Upf)
	if err != nil {
		return 0, fmt.Errorf("Slice %s unable to determine upf: %s", *slice.SliceId, err)
	}
	err = validateUpf(aUpf)
	if err != nil {
		return 0, fmt.Errorf("Slice %s Upf is invalid: %s", *slice.SliceId, err)
	}

	if aUpf.ConfigEndpoint == nil {
		return 0, fmt.Errorf("Slice %s UPF has no configuration endpoint", *slice.SliceId)
	}

	sc := &upfSliceConfig{
		SliceName: *slice.SliceId,
	}

	hasQos := false
	if slice.Mbr != nil {
		if slice.Mbr.Uplink != nil {
			sc.SliceQos.Uplink = *slice.Mbr.Uplink
			sc.SliceQos.UplinkBurst = DerefUint32Ptr(slice.Mbr.UplinkBurstSize, DefaultUplinkBurst)
			hasQos = true
		}
		if slice.Mbr.Downlink != nil {
			sc.SliceQos.Downlink = *slice.Mbr.Downlink
			sc.SliceQos.DownlinkBurst = DerefUint32Ptr(slice.Mbr.DownlinkBurstSize, DefaultDownlinkBurst)
			hasQos = true
		}
	}
	if hasQos {
		// Only specify units if we actually included a bitrate setting
		sc.SliceQos.Unit = aStr(DefaultBitrateUnit)
	}

	dgList, err := s.GetSliceDG(scope, slice)
	if err != nil {
		return 0, fmt.Errorf("Slice %s unable to determine dgList: %s", *slice.SliceId, err)
	}

	for _, dg := range dgList {
		ipd, err := s.GetIPDomain(scope, dg.IpDomain)
		if err != nil {
			return 0, fmt.Errorf("DeviceGroup %s failed to get IpDomain: %s", *dg.DeviceGroupId, err)
		}

		if ipd.Dnn != nil {
			ueRes := ueResourceInfo{Pool: *dg.DeviceGroupId,
				DNN: *ipd.Dnn}
			sc.UEResourceInfo = append(sc.UEResourceInfo, ueRes)
		}
	}

	if s.partialUpdateEnable && s.CacheCheck(CacheModelSliceUpf, *slice.SliceId, sc) {
		log.Infof("UPF Slice %s has not changed", *slice.SliceId)
		return 0, nil
	}

	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("Slice %s failed to marshal UPF JSON: %s", *slice.SliceId, err)
	}

	url := fmt.Sprintf("%s/v1/config/network-slices", *aUpf.ConfigEndpoint)
	err = s.pusher.PushUpdate(url, data)
	if err != nil {
		return 1, fmt.Errorf("slice %s failed to push UPF JSON: %s", *slice.SliceId, err)
	}

	s.CacheUpdate(CacheModelSliceUpf, *slice.SliceId, sc)

	return 0, nil
}
