// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package gnmi implements a gnmi server to mock a device with YANG models.
package synchronizer

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/openconfig/ygot/ygot"
	"net/http"
	"os"
	"time"

	models "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
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
	From uint64 `json:"from"`
	To   uint64 `json:"to"`
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
	Priority        *uint32        `json:"priority,omitempty"`
	Keys            SubscriberKeys `json:"keys"`
	ApnProfile      *string        `json:"selected-apn-profile,omitempty"`
	AccessProfile   []string       `json:"selected-access-profile,omitempty"`
	QosProfile      *string        `json:"selected-qos-profile,omitempty"`
	UpProfile       *string        `json:"selected-user-plane-profile,omitempty"`
	SecurityProfile *string        `json:"selected-security-profile,omitempty"`
}

type ApnProfile struct {
	ApnName      *string `json:"apn-name"`
	DnsPrimary   *string `json:"dns-primary"`
	DnsSecondary *string `json:"dns-secondary"`
	Mtu          *uint32 `json:"mtu"`
	GxEnabled    *bool   `json:"gx-enabled"`
	Network      string  `json:"network"`
	Usage        uint32  `json:"usage"`
}

type UpProfile struct {
	UserPlane     *string           `json:"user-plane"`
	AccessControl *string           `json:"access-control,omitempty"`
	AccessTags    map[string]string `json:"access-tags"`
	QosTags       map[string]string `json:"qos-tags"`
}

type QosArp struct {
	Priority                uint32 `json:"priority"`
	PreemptionCapability    uint32 `json:"pre-emption-capability"`
	PreemptionVulnerability uint32 `json:"pre-emption-vulnerability"`
}

type QosProfile struct {
	ApnAmbr []uint32 `json:"apn-ambr"`
	Qci     uint32   `json:"qci"`
	Arp     *QosArp  `json:"arp"`
}

type AccessProfile struct {
	Type   *string `json:"type"`
	Filter *string `json:"filter,omitempty"`
}

type SecurityProfile struct {
	Key *string `json:"key"`
	Opc *string `json:"opc"`
	Sqn *uint32 `json:"sqn"`
}

// On all of these, consider whether it is preferred to leave the item out if empty, or
// to emit an empty list.
type JsonConfig struct {
	SubscriberSelectionRules []SubscriberSelectionRule  `json:"subscriber-selection-rules,omitempty"`
	AccessProfiles           map[string]AccessProfile   `json:"access-profiles,omitempty"`
	ApnProfiles              map[string]ApnProfile      `json:"apn-profiles,omitempty"`
	QosProfiles              map[string]QosProfile      `json:"qos-profiles,omitempty"`
	UpProfiles               map[string]UpProfile       `json:"user-plane-profiles,omitempty"`
	SecurityProfiles         map[string]SecurityProfile `json:"security-profiles,omitempty"`
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

	log.Infof("Post returned status %s", resp.Status)

	if resp.StatusCode != 200 {
		return fmt.Errorf("Post returned error %s", resp.Status)
	}

	return nil
}

func (s *Synchronizer) SynchronizeDevice(config ygot.ValidatedGoStruct) error {
	device := config.(*models.Device)

	if device.Enterprise == nil {
		log.Info("No enteprises")
		return nil
	}

	if device.ConnectivityService == nil {
		log.Info("No connectivity services")
		return nil
	}

	for entId, ent := range device.Enterprise.Enterprise {
		if len(ent.ConnectivityService) == 0 {
			log.Info("Enterprise %s has no Connectivity Services", entId)
			// nothing to see here, move along.
			continue
		}
		for csId := range ent.ConnectivityService {
			cs, ok := device.ConnectivityService.ConnectivityService[csId]
			if !ok {
				return fmt.Errorf("Failed to find connectivity service %s", csId)
			}

			err := s.SynchronizeConnectivityService(device, ent, cs)
			if err != nil {
				// TODO: Think about this more -- if one fails then we end up aborting them all...
				return err
			}
		}
	}

	return nil
}

func (s *Synchronizer) SynchronizeConnectivityService(device *models.Device, ent *models.Enterprise_Enterprise_Enterprise, cs *models.ConnectivityService_ConnectivityService_ConnectivityService) error {
	_ = ent
	_ = cs

	jsonConfig := JsonConfig{}

	if device.Subscriber != nil {
		for _, ue := range device.Subscriber.Ue {
			keys := SubscriberKeys{}

			if (ue.Enterprise == nil) || (ent.Id == nil) {
				// The UE has no enterprise, or the enterprise has no Id
				continue
			}

			if *ue.Enterprise != *ent.Id {
				// The UE is for some other Enterprise than the one we're working on
				continue
			}

			if (ue.Enabled == nil) || (!*ue.Enabled) {
				continue
			}

			if ue.ServingPlmn != nil {
				keys.ServingPlmn = &SubscriberServingPlmn{
					Mcc: *ue.ServingPlmn.Mcc,
					Mnc: *ue.ServingPlmn.Mnc,
					Tac: *ue.ServingPlmn.Tac,
				}
			}

			if ue.ImsiRangeFrom != nil || ue.ImsiRangeTo != nil {
				// If we have one, then we require the other.
				if ue.ImsiRangeFrom == nil {
					return errors.New("ImsiRangeFrom is nil, but ImsiRangeTo is not")
				}
				if ue.ImsiRangeTo == nil {
					return errors.New("ImsiRangeTo is nil, but ImsiRangeFrom is not")
				}
				keys.ImsiRange = SubscriberImsiRange{
					From: *ue.ImsiRangeFrom,
					To:   *ue.ImsiRangeTo,
				}
			}

			if ue.RequestedApn != nil {
				keys.RequestedApn = *ue.RequestedApn
			}

			rule := SubscriberSelectionRule{
				Priority:        ue.Priority,
				Keys:            keys,
				ApnProfile:      ue.Profiles.ApnProfile,
				QosProfile:      ue.Profiles.QosProfile,
				UpProfile:       ue.Profiles.UpProfile,
				SecurityProfile: ue.Profiles.SecurityProfile,
			}

			for _, ap := range ue.Profiles.AccessProfile {
				if *ap.Allowed {
					rule.AccessProfile = append(rule.AccessProfile, *ap.AccessProfile)
				}
			}

			jsonConfig.SubscriberSelectionRules = append(jsonConfig.SubscriberSelectionRules, rule)
		}
	}

	if device.ApnProfile != nil {
		jsonConfig.ApnProfiles = make(map[string]ApnProfile)
		for _, apn := range device.ApnProfile.ApnProfile {
			profile := ApnProfile{
				ApnName:      apn.ApnName,
				DnsPrimary:   apn.DnsPrimary,
				DnsSecondary: apn.DnsSecondary,
				Mtu:          apn.Mtu,
				GxEnabled:    apn.GxEnabled,
				Network:      "lbo", // TODO: update modeling and revise
				Usage:        1,     // TODO: update modeling and revise
			}

			jsonConfig.ApnProfiles[*apn.Id] = profile
		}
	}

	if device.AccessProfile != nil {
		jsonConfig.AccessProfiles = make(map[string]AccessProfile)
		for _, access := range device.AccessProfile.AccessProfile {
			profile := AccessProfile{
				Type:   access.Type,
				Filter: access.Filter,
			}

			jsonConfig.AccessProfiles[*access.Id] = profile
		}
	}

	if device.QosProfile != nil {
		jsonConfig.QosProfiles = make(map[string]QosProfile)
		for _, qos := range device.QosProfile.QosProfile {
			arp := QosArp{ // TODO: hardcoded - fixme
				PreemptionCapability:    1,
				PreemptionVulnerability: 1,
			}
			profile := QosProfile{
				ApnAmbr: []uint32{*qos.ApnAmbr.Downlink, *qos.ApnAmbr.Uplink},
				Qci:     9, // TODO: hardcoded - fixme
				Arp:     &arp,
			}

			jsonConfig.QosProfiles[*qos.Id] = profile
		}
	}

	if device.UpProfile != nil {
		jsonConfig.UpProfiles = make(map[string]UpProfile)
		for _, up := range device.UpProfile.UpProfile {
			profile := UpProfile{
				UserPlane:     up.UserPlane,
				AccessControl: up.AccessControl,
				AccessTags:    map[string]string{"tag1": "ACC"}, // TODO: update modeling and revise
				QosTags:       map[string]string{"tag1": "BW"},  // TODO: update modeling and revise
			}

			jsonConfig.UpProfiles[*up.Id] = profile
		}
	}

	if device.SecurityProfile != nil {
		jsonConfig.SecurityProfiles = make(map[string]SecurityProfile)
		for _, sp := range device.SecurityProfile.SecurityProfile {
			profile := SecurityProfile{
				Key: sp.Key,
				Opc: sp.Opc,
				Sqn: sp.Sqn,
			}

			jsonConfig.SecurityProfiles[*sp.Id] = profile
		}
	}

	data, err := json.MarshalIndent(jsonConfig, "", "  ")
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

	if s.postEnable {
		if cs.SpgwcEndpoint != nil {
			log.Infof("Posting to %s", *cs.SpgwcEndpoint)
			err := s.Post(*cs.SpgwcEndpoint, data)
			if err != nil {
				return err
			}
		}

		if cs.HssEndpoint != nil {
			log.Infof("Posting to %s", *cs.HssEndpoint)
			err := s.Post(*cs.HssEndpoint, data)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
