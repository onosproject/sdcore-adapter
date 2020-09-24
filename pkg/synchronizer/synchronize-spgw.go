// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package synchronizer

import (
	"bytes"
	"encoding/json"
	"github.com/openconfig/ygot/ygot"
	"net/http"
	"os"
	"strings"
	"time"

	models "github.com/onosproject/config-models/modelplugin/aether-1.0.0/aether_1_0_0"
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

type SubscriberImsiRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type SubscriberServingPlmn struct {
	Mcc uint32 `json:"mcc"`
	Mnc uint32 `json:"mnc"`
	Tac uint32 `json:"tac"`
}

type SubscriberKeys struct {
	ImsiRange    SubscriberImsiRange    `json:"imsi-range,omitempty"`
	ServingPlmn  *SubscriberServingPlmn `json:"serving-plmn,omitempty"`
	RequestedApn string                 `json:"requested-apn,omitempty"`
}

type SubscriberSelectionRule struct {
	Priority      *uint32        `json:"priority,omitempty"`
	Keys          SubscriberKeys `json:"keys"`
	ApnProfile    *string        `json:"selected-apn-profile,omitempty"`
	AccessProfile []string       `json:"selected-access-profile,omitempty"`
	QosProfile    *string        `json:"selected-qos-profile,omitempty"`
	UpProfile     *string        `json:"selected-user-plane-profile,omitempty"`
}

type ApnProfile struct {
	ApnName      *string `json:"apn-name"`
	DnsPrimary   *string `json:"dns-primary"`
	DnsSecondary *string `json:"dns-secondary"`
	Mtu          *uint32 `json:"mtu"`
	GxEnabled    *bool   `json:"gx-enabled"`
}

type UpProfile struct {
	UserPlane     *string `json:"user-plane"`
	AccessControl *string `json:"access-control"`
}

type QosProfile struct {
	ApnAmbr []uint32 `json:"apn-ambr"`
}

type AccessProfile struct {
	Type   *string `json:"type"`
	Filter *string `json:"filter"`
}

// On all of these, consider whether it is preferred to leave the item out if empty, or
// to emit an empty list.
type SpgwConfig struct {
	SubscriberSelectionRules []SubscriberSelectionRule `json:"subscriber-selection-rules,omitempty"`
	AccessProfiles           map[string]AccessProfile  `json:"access-profiles,omitempty"`
	ApnProfiles              map[string]ApnProfile     `json:"apn-profiles,omitempty"`
	QosProfiles              map[string]QosProfile     `json:"qos-profiles,omitempty"`
	UpProfiles               map[string]UpProfile      `json:"user-plane-profiles,omitempty"`
}

func ConvertImsiRange(s string) SubscriberImsiRange {
	// TODO: Bug in onos-config causes strings that are all digits to be
	// converted into integers. I've been using the workaround of substituting
	// an "e" for the first "3" in the IMSI.
	s = strings.Replace(s, "e", "3", -1)

	var imsiRange SubscriberImsiRange
	if strings.Contains(s, "-") {
		parts := strings.SplitN(s, "-", 2)
		imsiRange = SubscriberImsiRange{
			From: parts[0],
			To:   parts[1],
		}
	} else {
		imsiRange = SubscriberImsiRange{
			From: s,
			To:   s,
		}
	}
	return imsiRange
}

func (s *Synchronizer) Post(endpoint string, data []byte) error {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	resp, err := client.Post(
		endpoint,
		"application/json",
		bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (s *Synchronizer) SynchronizeSpgw(config ygot.ValidatedGoStruct) error {
	device := config.(*models.Device)

	spgwConfig := SpgwConfig{}

	if device.Subscriber != nil {
		for _, ue := range device.Subscriber.Ue {
			var servingPlmn *SubscriberServingPlmn
			if ue.ServingPlmn != nil {
				servingPlmn = &SubscriberServingPlmn{
					Mcc: *ue.ServingPlmn.Mcc,
					Mnc: *ue.ServingPlmn.Mnc,
					Tac: *ue.ServingPlmn.Tac,
				}
			}

			rule := SubscriberSelectionRule{
				Priority: ue.Priority,
				Keys: SubscriberKeys{
					ImsiRange:    ConvertImsiRange(*ue.Ueid),
					ServingPlmn:  servingPlmn,
					RequestedApn: *ue.RequestedApn,
				},
				ApnProfile: ue.Profiles.ApnProfile,
				QosProfile: ue.Profiles.QosProfile,
				UpProfile:  ue.Profiles.UpProfile,
			}

			for _, ap := range ue.Profiles.AccessProfile {
				if *ap.Allowed {
					rule.AccessProfile = append(rule.AccessProfile, *ap.AccessProfile)
				}
			}

			spgwConfig.SubscriberSelectionRules = append(spgwConfig.SubscriberSelectionRules, rule)
		}
	}

	if device.ApnProfile != nil {
		spgwConfig.ApnProfiles = make(map[string]ApnProfile)
		for _, apn := range device.ApnProfile.ApnProfile {
			profile := ApnProfile{
				ApnName:      apn.ApnName,
				DnsPrimary:   apn.DnsPrimary,
				DnsSecondary: apn.DnsSecondary,
				Mtu:          apn.Mtu,
				GxEnabled:    apn.GxEnabled,
			}

			spgwConfig.ApnProfiles[*apn.Id] = profile
		}
	}

	if device.AccessProfile != nil {
		spgwConfig.AccessProfiles = make(map[string]AccessProfile)
		for _, access := range device.AccessProfile.AccessProfile {
			profile := AccessProfile{
				Type:   access.Type,
				Filter: access.Filter,
			}

			spgwConfig.AccessProfiles[*access.Id] = profile
		}
	}

	if device.QosProfile != nil {
		spgwConfig.QosProfiles = make(map[string]QosProfile)
		for _, qos := range device.QosProfile.QosProfile {
			profile := QosProfile{
				ApnAmbr: []uint32{*qos.ApnAmbr.Downlink, *qos.ApnAmbr.Uplink},
			}

			spgwConfig.QosProfiles[*qos.Id] = profile
		}
	}

	if device.UpProfile != nil {
		spgwConfig.UpProfiles = make(map[string]UpProfile)
		for _, up := range device.UpProfile.UpProfile {
			profile := UpProfile{
				UserPlane:     up.UserPlane,
				AccessControl: up.AccessControl,
			}

			spgwConfig.UpProfiles[*up.Id] = profile
		}
	}

	//log.Infof("spgwConfig %v", spgwConfig)

	data, err := json.MarshalIndent(spgwConfig, "", "  ")
	if err != nil {
		return err
	}

	log.Infof("Emit: %v", string(data))

	if s.outputFileName != "" {
		log.Infof("Writing %s", s.outputFileName)
		file, err := os.OpenFile(
			s.outputFileName,
			os.O_WRONLY|os.O_TRUNC|os.O_CREATE,
			0666,
		)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = file.Write(data)
		if err != nil {
			return err
		}
	}

	if s.spgwEndpoint != "" {
		log.Infof("Posting to %s", s.spgwEndpoint)
		s.Post(s.spgwEndpoint, data)

	}

	return nil
}
