// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizer

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type plmnId struct {
        Mcc string `json:"mcc,omitempty"`
        Mnc string `json:"mnc,omitempty"`
}

type sliceId struct {
        Sst    int `json:"sst,omitempty"`
        Sd     string `json:"sd,omitempty"`
        PlmnId plmnId `json:"plmnId"`
}

type scope struct {
	SliceId sliceId `json:"sliceId"`
}

type rrmInfo struct {
        RRMMin       uint8 `json:"guaDlThptPerSlice,omitempty"`
        RRMMax       uint8 `json:"maxDlThptPerSlice,omitempty"`
        RRMDedicated uint8 `json:"maxDlThptPerUe,omitempty"`
}

type xappSliceConfig struct {
        Scope   scope `json:"scope"`
        RrmInfo   rrmInfo `json:"sliceSlaObjectives"`
}

// SynchronizeSliceUPF synchronizes the VCSes to the UPF
// Return a count of push-related errors
func (s *Synchronizer) SynchronizeSliceXAPP(scope *AetherScope, slice *Slice) (int, error) {
	if scope.Site.ConnectivityService.Ran_5GService.XappEndpoint == nil || *scope.Site.ConnectivityService.Ran_5GService.XappEndpoint == "" {
		return 1, fmt.Errorf("Site %s Ran_5GService has no endpoint", *scope.Site.SiteId)
	}

	sc := &xappSliceConfig{
		//SliceName: *slice.SliceId,
	}
	intVar, err := strconv.Atoi(*slice.Sst)
        sc.Scope.SliceId.Sst = intVar
        sc.Scope.SliceId.Sd = *slice.Sd
        sc.Scope.SliceId.PlmnId.Mcc = *scope.Site.ImsiDefinition.Mcc
        sc.Scope.SliceId.PlmnId.Mnc = *scope.Site.ImsiDefinition.Mnc
	sc.RrmInfo.RRMMin = *slice.Xapp.RrmMin
	sc.RrmInfo.RRMMax = *slice.Xapp.RrmMax
	sc.RrmInfo.RRMDedicated = *slice.Xapp.RrmDedicated

	if s.partialUpdateEnable && s.CacheCheck(CacheModelSliceXapp, *slice.SliceId, sc) {
		log.Infof("UPF Slice %s has not changed", *slice.SliceId)
		return 0, nil
	}

	data, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return 0, fmt.Errorf("Slice %s failed to marshal XAPP JSON: %s", *slice.SliceId, err)
	}

	url := fmt.Sprintf("%s/policytypes/ORAN_SliceSLATarget_1.0.0/policies/1", *scope.Site.ConnectivityService.Ran_5GService.XappEndpoint)
	err = s.pusher.PushUpdate(url, data)
	if err != nil {
		return 1, fmt.Errorf("slice %s failed to push XAPP JSON: %s", *slice.SliceId, err)
	}

	s.CacheUpdate(CacheModelSliceXapp, *slice.SliceId, sc)

	return 0, nil
}
