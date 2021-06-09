// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizer implements a synchronizer for converting sdcore gnmi to json
package synchronizerv2

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/openconfig/ygot/ygot"
	"net/http"
	"os"
	"time"

	models "github.com/onosproject/config-models/modelplugin/aether-2.1.0/aether_2_1_0"
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
	ImsiRange    *SubscriberImsiRange   `json:"imsi-range,omitempty"`
	ServingPlmn  *SubscriberServingPlmn `json:"serving-plmn,omitempty"`
	RequestedApn string                 `json:"requested-apn,omitempty"`
	MatchAll     *bool                  `json:"match-all,omitempty"`
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
	DnsPrimary   *string `json:"dns_primary"`
	DnsSecondary *string `json:"dns_secondary"`
	Mtu          *uint32 `json:"mtu"`
	GxEnabled    *bool   `json:"gx_enabled"`
	Network      string  `json:"network"`
	Usage        uint32  `json:"usage"`
	GxApn        *string `json:"gx_apn,omitempty"`
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

type ServiceGroup struct {
	DefaultActivateService *string  `json:"default-activate-service,omitempty"`
	OnDemandService        []string `json:"on-demand-service,omitempty"`
}

type Service struct {
	Qci                  *uint32  `json:"qci,omitempty"`
	Arp                  *uint32  `json:"arp,omitempty"`
	AMBR_UL              *uint32  `json:"AMBR_UL,omitempty"`
	AMBR_DL              *uint32  `json:"AMBR_DL,omitempty"`
	Rules                []string `json:"service-activation-rules"`
	ActivateConditions   []string `json:"activate-conditions,omitempty"`
	DeactivateConditions []string `json:"deactivate-conditions,omitempty"`
	DeactivateActions    []string `json:"deactivate-acrionts,omitempty"`
}

type RuleDefinitionQosArp struct {
	Priority                uint32 `json:"Priority-Level,omitempty"`
	PreemptionCapability    uint32 `json:"Pre-emption-Capability,omitempty"`
	PreemptionVulnerability uint32 `json:"Pre-emption-Vulnerability,omitempty"`
}

type RuleDefinitionQos struct {
	Qci       *uint32               `json:"QoS-Class-Identifier,omitempty"`
	MBRUL     *uint32               `json:"Max-Requested-Bandwidth-UL,omitempty"`
	MBRDL     *uint32               `json:"Max-Requested-Bandwidth-DL,omitempty"`
	GBUL      *uint32               `json:"Guaranteed-Bitrate-UL,omitempty"`
	GBDL      *uint32               `json:"Guaranteed-Bitrate-DL,omitempty"`
	Arp       *RuleDefinitionQosArp `json:"Allocation-Retention-Policy,omitempty"`
	APNAMBRUL *uint32               `json:"APN-Aggregate-Max-Bitrate-UL,omitempty"`
	APNAMBRDL *uint32               `json:"APN-Aggregate-Max-Bitrate-DL,omitempty"`
}

type FlowInformation struct {
	FlowDesc string `json:"Flow-Description"`
}

type RuleDefinition struct {
	ChargingRuleName *string            `json:"Charging-Rule-Name,omitempty"`
	QosInformation   *RuleDefinitionQos `json:"QoS-Information,omitempty"`
	FlowInformation  *FlowInformation   `json:"Flow-Information,omitempty"`
}

type Rule struct {
	Definition *RuleDefinition `json:"definition,omitempty"`
}

type PoliciesStruct struct {
	ServiceGroups map[string]ServiceGroup `json:"service-groups,omitempty"`
	Services      map[string]Service      `json:"services,omitempty"`
	Rules         map[string]Rule         `json:"rules,omitempty"`
}

/* SD-Core JSON Config
 *
 * All SD-Core components are able to accept the full config, though some
 * components only pay attention to certain parts of it:
 *     SPGWC: SubscriberSelectionRules,AccessProfiles,ApnProfiles,QosProfiles,UPProfiles
 *     HSS: SubscriberSelectionRules,SecurityProfiles
 *     PCRF: Policies
 * The excess can be omitted when posting for performance improvement
 *    (...and it also works around the large packet size bug)
 */
type JsonConfig struct {
	SubscriberSelectionRules []SubscriberSelectionRule  `json:"subscriber-selection-rules,omitempty"`
	AccessProfiles           map[string]AccessProfile   `json:"access-profiles,omitempty"`
	ApnProfiles              map[string]ApnProfile      `json:"apn-profiles,omitempty"`
	QosProfiles              map[string]QosProfile      `json:"qos-profiles,omitempty"`
	UpProfiles               map[string]UpProfile       `json:"user-plane-profiles,omitempty"`
	SecurityProfiles         map[string]SecurityProfile `json:"security-profiles,omitempty"`
	Policies                 *PoliciesStruct            `json:"policies,omitempty"`
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

	// For a given ConnectivityService, we want to know the list of Enterprises
	// that use it. Precompute this so we can pass a list of valid Enterprises
	// along to SynchronizeConnectivityService.
	csEntMap := map[string]map[string]bool{}
	for entId, ent := range device.Enterprise.Enterprise {
		for csId := range ent.ConnectivityService {
			m, okay := csEntMap[csId]
			if !okay {
				m = map[string]bool{}
				csEntMap[csId] = m
			}
			m[entId] = true
		}
	}

	errors := []error{}
	for csId, cs := range device.ConnectivityService.ConnectivityService {
		// Get the list of valid Enterprises for this CS.
		// Note: This could return an empty map if there is a CS that no
		//   enterprises are linked to . In that case, we can still push models
		//   that are not directly related to an enterprise, such as profiles.
		m := csEntMap[csId]

		tStart := time.Now()
		synchronizationTotal.WithLabelValues(csId).Inc()

		err := s.SynchronizeConnectivityService(device, cs, m)
		if err != nil {
			synchronizationFailedTotal.WithLabelValues(csId).Inc()
			// If there are errors, then build a list of them and continue to try
			// to synchronize other connectivity services.
			errors = append(errors, err)
		} else {
			synchronizationDuration.WithLabelValues(csId).Observe(time.Since(tStart).Seconds())
		}
	}

	if len(errors) == 0 {
		return nil
	} else {
		return fmt.Errorf("synchronization errors: %v", errors)
	}
}

func (s *Synchronizer) SynchronizePCRF(device *models.Device) (*PoliciesStruct, error) {
	policies := PoliciesStruct{}

	if device.ServiceGroup != nil {
		policies.ServiceGroups = make(map[string]ServiceGroup)

		for _, sg := range device.ServiceGroup.ServiceGroup {
			jsg := ServiceGroup{}
			for _, s := range sg.ServicePolicies {
				if *s.Kind == "default" {
					jsg.DefaultActivateService = s.ServicePolicy
				} else { // on-demand
					jsg.OnDemandService = append(jsg.OnDemandService, *s.ServicePolicy)
				}
			}
			policies.ServiceGroups[*sg.Id] = jsg
		}
	}

	if device.ServicePolicy != nil {
		policies.Services = make(map[string]Service)

		for _, sp := range device.ServicePolicy.ServicePolicy {
			jsp := Service{Qci: sp.Qci,
				Arp: sp.Arp}

			if sp.Ambr != nil {
				jsp.AMBR_UL = sp.Ambr.Uplink
				jsp.AMBR_DL = sp.Ambr.Downlink
			}

			for _, rule := range sp.Rules {
				if *rule.Enabled {
					jsp.Rules = append(jsp.Rules, *rule.Rule)
				}
			}

			policies.Services[*sp.Id] = jsp
		}
	}

	if device.ServiceRule != nil {
		policies.Rules = make(map[string]Rule)

		for _, rule := range device.ServiceRule.ServiceRule {
			def := RuleDefinition{
				ChargingRuleName: rule.ChargingRuleName,
			}

			if rule.Qos != nil {
				qos := RuleDefinitionQos{
					Qci: rule.Qos.Qci,
				}

				if rule.Qos.MaximumRequestedBandwidth != nil {
					qos.MBRUL = rule.Qos.MaximumRequestedBandwidth.Uplink
					qos.MBRDL = rule.Qos.MaximumRequestedBandwidth.Downlink
				}

				if rule.Qos.GuaranteedBitrate != nil {
					qos.GBUL = rule.Qos.GuaranteedBitrate.Uplink
					qos.GBDL = rule.Qos.GuaranteedBitrate.Downlink
				}

				if rule.Qos.AggregateMaximumBitrate != nil {
					qos.APNAMBRUL = rule.Qos.AggregateMaximumBitrate.Uplink
					qos.APNAMBRDL = rule.Qos.AggregateMaximumBitrate.Downlink
				}

				if rule.Qos.Arp != nil {
					arp := RuleDefinitionQosArp{
						Priority:                *rule.Qos.Arp.Priority,
						PreemptionCapability:    boolToUint32(*rule.Qos.Arp.PreemptionCapability),
						PreemptionVulnerability: boolToUint32(*rule.Qos.Arp.PreemptionVulnerability),
					}
					qos.Arp = &arp
				}

				def.QosInformation = &qos
			}

			if rule.Flow != nil {
				jflow := FlowInformation{
					FlowDesc: *rule.Flow.Specification,
				}

				def.FlowInformation = &jflow
			}

			jrule := Rule{Definition: &def}
			policies.Rules[*rule.Id] = jrule
		}
	}

	return &policies, nil
}

type FilterDataFunc func(*JsonConfig) *JsonConfig

// Return the portion of the JsonConfig relevant to SPGWC
func FilterConfigSPGWC(src *JsonConfig) *JsonConfig {
	dest := JsonConfig{
		SubscriberSelectionRules: src.SubscriberSelectionRules,
		AccessProfiles:           src.AccessProfiles,
		ApnProfiles:              src.ApnProfiles,
		QosProfiles:              src.QosProfiles,
		UpProfiles:               src.UpProfiles,
	}
	return &dest
}

// Return the portion of the JsonConfig relevant to HSS
func FilterConfigHSS(src *JsonConfig) *JsonConfig {
	dest := JsonConfig{
		SubscriberSelectionRules: src.SubscriberSelectionRules,
		SecurityProfiles:         src.SecurityProfiles,
		// HSS also seems to need these, or it will ignore the JSON
		ApnProfiles: src.ApnProfiles,
		QosProfiles: src.QosProfiles,
	}
	return &dest
}

// Return the portion of the JsonConfig relevant to PCRF
func FilterConfigPCRF(src *JsonConfig) *JsonConfig {
	dest := JsonConfig{
		Policies: src.Policies,
	}
	return &dest
}

func (s *Synchronizer) PostData(name string, endpoint string, filter FilterDataFunc, jsonConfig *JsonConfig) error {
	var filteredConfig *JsonConfig
	if filter == nil {
		filteredConfig = jsonConfig

	} else {
		filteredConfig = filter(jsonConfig)
	}

	data, err := json.MarshalIndent(*filteredConfig, "", "  ")
	if err != nil {
		return err
	}

	log.Infof("Data to %s: %v", name, string(data))

	log.Infof("Posting %d bytes to %s at %s", len(data), name, endpoint)
	err = s.Post(endpoint, data)
	if err != nil {
		return err
	}

	return nil
}

func (s *Synchronizer) SynchronizeConnectivityService(device *models.Device, cs *models.ConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) error {
	jsonConfig := JsonConfig{}

	log.Infof("Synchronizing Connectivity Service %s", *cs.Id)

	if device.Subscriber != nil {
		for _, ue := range device.Subscriber.Ue {
			keys := SubscriberKeys{}

			if ue.Enterprise == nil {
				// The UE has no enterprise, or the enterprise has no Id
				log.Infof("UE %s has no enterprise", *ue.Id)
				continue
			}

			if ue.Profiles == nil {
				// Require there to be at least some profiles before we'll consider it
				log.Infof("UE %s has no profiles", *ue.Id)
				continue
			}

			_, okay := validEnterpriseIds[*ue.Enterprise]
			if !okay {
				// The UE is for some other CS than the one we're working on
				log.Infof("UE %s is not for connectivity service %s", *ue.Id, *cs.Id)
				continue
			}

			if (ue.Enabled == nil) || (!*ue.Enabled) {
				log.Infof("UE %s is not enabled", *ue.Id)
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
				keys.ImsiRange = &SubscriberImsiRange{
					From: *ue.ImsiRangeFrom,
					To:   *ue.ImsiRangeTo,
				}
			}

			if ue.RequestedApn != nil {
				keys.RequestedApn = *ue.RequestedApn
			}

			// if no keys are specified, then emit the match-all rule
			if (ue.ServingPlmn == nil) && (ue.ImsiRangeFrom == nil) && (ue.ImsiRangeTo == nil) && (ue.RequestedApn == nil) {
				matchAll := true
				keys.MatchAll = &matchAll
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
				GxApn:        apn.ServiceGroup,
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
			arp := QosArp{
				Priority:                0, // default
				PreemptionCapability:    1, // default
				PreemptionVulnerability: 1, // default
			}
			if qos.Arp != nil {
				if qos.Arp.Priority != nil {
					arp.Priority = *qos.Arp.Priority
				}
				if qos.Arp.PreemptionCapability != nil {
					arp.PreemptionCapability = boolToUint32(*qos.Arp.PreemptionCapability)
				}
				if qos.Arp.PreemptionVulnerability != nil {
					arp.PreemptionVulnerability = boolToUint32(*qos.Arp.PreemptionVulnerability)
				}
			}

			profile := QosProfile{
				Qci: 9, // default
				Arp: &arp,
			}

			if qos.ApnAmbr != nil {
				profile.ApnAmbr = []uint32{*qos.ApnAmbr.Downlink, *qos.ApnAmbr.Uplink}
			}

			if qos.Qci != nil {
				profile.Qci = *qos.Qci
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

	if (device.ServicePolicy != nil) || (device.ServiceRule != nil) || (device.ServiceGroup != nil) {
		var err error
		jsonConfig.Policies, err = s.SynchronizePCRF(device)
		if err != nil {
			return err
		}
	}

	synchronizationResourceTotal.WithLabelValues(*cs.Id, "subscriber").Set(float64(len(jsonConfig.SubscriberSelectionRules)))
	synchronizationResourceTotal.WithLabelValues(*cs.Id, "apn-profile").Set(float64(len(jsonConfig.ApnProfiles)))
	synchronizationResourceTotal.WithLabelValues(*cs.Id, "access-profile").Set(float64(len(jsonConfig.AccessProfiles)))
	synchronizationResourceTotal.WithLabelValues(*cs.Id, "qos-profile").Set(float64(len(jsonConfig.QosProfiles)))
	synchronizationResourceTotal.WithLabelValues(*cs.Id, "up-profile").Set(float64(len(jsonConfig.UpProfiles)))
	synchronizationResourceTotal.WithLabelValues(*cs.Id, "security-profile").Set(float64(len(jsonConfig.SecurityProfiles)))

	if s.outputFileName != "" {
		data, err := json.MarshalIndent(jsonConfig, "", "  ")
		if err != nil {
			return err
		}

		log.Infof("Writing to file: %v", string(data))

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
			err := s.PostData("SPGWC", *cs.SpgwcEndpoint, FilterConfigSPGWC, &jsonConfig)
			if err != nil {
				return err
			}
		}

		if cs.HssEndpoint != nil {
			err := s.PostData("HSS", *cs.HssEndpoint, FilterConfigHSS, &jsonConfig)
			if err != nil {
				return err
			}
		}

		if cs.PcrfEndpoint != nil {
			err := s.PostData("PCRF", *cs.PcrfEndpoint, FilterConfigPCRF, &jsonConfig)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
