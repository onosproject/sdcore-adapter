// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

// Package synchronizerv2 implements a synchronizer for converting sdcore gnmi to json
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
	"github.com/onosproject/sdcore-adapter/pkg/synchronizer"
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

type subscriberImsiRange struct {
	From uint64 `json:"from"`
	To   uint64 `json:"to"`
}

type subscriberServingPlmn struct {
	Mcc uint32 `json:"mcc"`
	Mnc uint32 `json:"mnc"`
	Tac uint32 `json:"tac"`
}

type subscriberKeys struct {
	ImsiRange    *subscriberImsiRange   `json:"imsi-range,omitempty"`
	ServingPlmn  *subscriberServingPlmn `json:"serving-plmn,omitempty"`
	RequestedApn string                 `json:"requested-apn,omitempty"`
	MatchAll     *bool                  `json:"match-all,omitempty"`
}

type subscriberSelectionRule struct {
	Priority        *uint32        `json:"priority,omitempty"`
	Keys            subscriberKeys `json:"keys"`
	ApnProfile      *string        `json:"selected-apn-profile,omitempty"`
	AccessProfile   []string       `json:"selected-access-profile,omitempty"`
	QosProfile      *string        `json:"selected-qos-profile,omitempty"`
	UpProfile       *string        `json:"selected-user-plane-profile,omitempty"`
	SecurityProfile *string        `json:"selected-security-profile,omitempty"`
}

type apnProfile struct {
	ApnName      *string `json:"apn-name"`
	DNSPrimary   *string `json:"dns_primary"`
	DNSSecondary *string `json:"dns_secondary"`
	Mtu          *uint32 `json:"mtu"`
	GxEnabled    *bool   `json:"gx_enabled"`
	Network      string  `json:"network"`
	Usage        uint32  `json:"usage"`
	GxApn        *string `json:"gx_apn,omitempty"`
}

type upProfile struct {
	UserPlane     *string           `json:"user-plane"`
	AccessControl *string           `json:"access-control,omitempty"`
	AccessTags    map[string]string `json:"access-tags"`
	QosTags       map[string]string `json:"qos-tags"`
}

type qosArp struct {
	Priority                uint32 `json:"priority"`
	PreemptionCapability    uint32 `json:"pre-emption-capability"`
	PreemptionVulnerability uint32 `json:"pre-emption-vulnerability"`
}

type qosProfile struct {
	ApnAmbr []uint32 `json:"apn-ambr"`
	Qci     uint32   `json:"qci"`
	Arp     *qosArp  `json:"arp"`
}

type accessProfile struct {
	Type   *string `json:"type"`
	Filter *string `json:"filter,omitempty"`
}

type securityProfile struct {
	Key *string `json:"key"`
	Opc *string `json:"opc"`
	Sqn *uint32 `json:"sqn"`
}

type serviceGroup struct {
	DefaultActivateService *string  `json:"default-activate-service,omitempty"`
	OnDemandService        []string `json:"on-demand-service,omitempty"`
}

type service struct {
	Qci                  *uint32  `json:"qci,omitempty"`
	Arp                  *uint32  `json:"arp,omitempty"`
	AMBR_UL              *uint32  `json:"AMBR_UL,omitempty"` //nolint
	AMBR_DL              *uint32  `json:"AMBR_DL,omitempty"` //nolint
	Rules                []string `json:"service-activation-rules"`
	ActivateConditions   []string `json:"activate-conditions,omitempty"`
	DeactivateConditions []string `json:"deactivate-conditions,omitempty"`
	DeactivateActions    []string `json:"deactivate-acrionts,omitempty"`
}

type ruleDefinitionQosArp struct {
	Priority                uint32 `json:"Priority-Level,omitempty"`
	PreemptionCapability    uint32 `json:"Pre-emption-Capability,omitempty"`
	PreemptionVulnerability uint32 `json:"Pre-emption-Vulnerability,omitempty"`
}

type ruleDefinitionQos struct {
	Qci       *uint32               `json:"QoS-Class-Identifier,omitempty"`
	MBRUL     *uint32               `json:"Max-Requested-Bandwidth-UL,omitempty"`
	MBRDL     *uint32               `json:"Max-Requested-Bandwidth-DL,omitempty"`
	GBUL      *uint32               `json:"Guaranteed-Bitrate-UL,omitempty"`
	GBDL      *uint32               `json:"Guaranteed-Bitrate-DL,omitempty"`
	Arp       *ruleDefinitionQosArp `json:"Allocation-Retention-Policy,omitempty"`
	APNAMBRUL *uint32               `json:"APN-Aggregate-Max-Bitrate-UL,omitempty"`
	APNAMBRDL *uint32               `json:"APN-Aggregate-Max-Bitrate-DL,omitempty"`
}

type flowInformation struct {
	FlowDesc string `json:"Flow-Description"`
}

type ruleDefinition struct {
	ChargingRuleName *string            `json:"Charging-Rule-Name,omitempty"`
	QosInformation   *ruleDefinitionQos `json:"QoS-Information,omitempty"`
	FlowInformation  *flowInformation   `json:"Flow-Information,omitempty"`
}

type ruleStruct struct {
	Definition *ruleDefinition `json:"definition,omitempty"`
}

type policiesStruct struct {
	ServiceGroups map[string]serviceGroup `json:"service-groups,omitempty"`
	Services      map[string]service      `json:"services,omitempty"`
	Rules         map[string]ruleStruct   `json:"rules,omitempty"`
}

/* jsonConfig: SD-Core JSON Config
 *
 * All SD-Core components are able to accept the full config, though some
 * components only pay attention to certain parts of it:
 *     SPGWC: SubscriberSelectionRules,AccessProfiles,ApnProfiles,QosProfiles,UPProfiles
 *     HSS: SubscriberSelectionRules,SecurityProfiles
 *     PCRF: Policies
 * The excess can be omitted when posting for performance improvement
 *    (...and it also works around the large packet size bug)
 */
type jsonConfig struct {
	SubscriberSelectionRules []subscriberSelectionRule  `json:"subscriber-selection-rules,omitempty"`
	AccessProfiles           map[string]accessProfile   `json:"access-profiles,omitempty"`
	ApnProfiles              map[string]apnProfile      `json:"apn-profiles,omitempty"`
	QosProfiles              map[string]qosProfile      `json:"qos-profiles,omitempty"`
	UpProfiles               map[string]upProfile       `json:"user-plane-profiles,omitempty"`
	SecurityProfiles         map[string]securityProfile `json:"security-profiles,omitempty"`
	Policies                 *policiesStruct            `json:"policies,omitempty"`
}

// Post to underlying service
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

// SynchronizeDevice synchronizes the device to the underlying service
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
	for endID, ent := range device.Enterprise.Enterprise {
		for csID := range ent.ConnectivityService {
			m, okay := csEntMap[csID]
			if !okay {
				m = map[string]bool{}
				csEntMap[csID] = m
			}
			m[endID] = true
		}
	}

	errors := []error{}
	for csID, cs := range device.ConnectivityService.ConnectivityService {
		// Get the list of valid Enterprises for this CS.
		// Note: This could return an empty map if there is a CS that no
		//   enterprises are linked to . In that case, we can still push models
		//   that are not directly related to an enterprise, such as profiles.
		m := csEntMap[csID]

		tStart := time.Now()
		synchronizer.KpiSynchronizationTotal.WithLabelValues(csID).Inc()

		err := s.SynchronizeConnectivityService(device, cs, m)
		if err != nil {
			synchronizer.KpiSynchronizationFailedTotal.WithLabelValues(csID).Inc()
			// If there are errors, then build a list of them and continue to try
			// to synchronize other connectivity services.
			errors = append(errors, err)
		} else {
			synchronizer.KpiSynchronizationDuration.WithLabelValues(csID).Observe(time.Since(tStart).Seconds())
		}
	}

	if len(errors) == 0 {
		return nil
	}

	return fmt.Errorf("synchronization errors: %v", errors)
}

// synchronizePCRF synchronizes the PCRF service
func (s *Synchronizer) synchronizePCRF(device *models.Device) (*policiesStruct, error) {
	policies := policiesStruct{}

	if device.ServiceGroup != nil {
		policies.ServiceGroups = make(map[string]serviceGroup)

		for _, sg := range device.ServiceGroup.ServiceGroup {
			jsg := serviceGroup{}
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
		policies.Services = make(map[string]service)

		for _, sp := range device.ServicePolicy.ServicePolicy {
			jsp := service{Qci: sp.Qci,
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
		policies.Rules = make(map[string]ruleStruct)

		for _, rule := range device.ServiceRule.ServiceRule {
			def := ruleDefinition{
				ChargingRuleName: rule.ChargingRuleName,
			}

			if rule.Qos != nil {
				qos := ruleDefinitionQos{
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
					arp := ruleDefinitionQosArp{
						Priority:                *rule.Qos.Arp.Priority,
						PreemptionCapability:    synchronizer.BoolToUint32(*rule.Qos.Arp.PreemptionCapability),
						PreemptionVulnerability: synchronizer.BoolToUint32(*rule.Qos.Arp.PreemptionVulnerability),
					}
					qos.Arp = &arp
				}

				def.QosInformation = &qos
			}

			if rule.Flow != nil {
				jflow := flowInformation{
					FlowDesc: *rule.Flow.Specification,
				}

				def.FlowInformation = &jflow
			}

			jrule := ruleStruct{Definition: &def}
			policies.Rules[*rule.Id] = jrule
		}
	}

	return &policies, nil
}

// FilterDataFunc is a function for filtering json config based on type of service
type FilterDataFunc func(*jsonConfig) *jsonConfig

// FilterConfigSPGWC returns the portion of the JsonConfig relevant to SPGWC
func filterConfigSPGWC(src *jsonConfig) *jsonConfig {
	dest := jsonConfig{
		SubscriberSelectionRules: src.SubscriberSelectionRules,
		AccessProfiles:           src.AccessProfiles,
		ApnProfiles:              src.ApnProfiles,
		QosProfiles:              src.QosProfiles,
		UpProfiles:               src.UpProfiles,
	}
	return &dest
}

// FilterConfigHSS returns the portion of the JsonConfig relevant to HSS
func filterConfigHSS(src *jsonConfig) *jsonConfig {
	dest := jsonConfig{
		SubscriberSelectionRules: src.SubscriberSelectionRules,
		SecurityProfiles:         src.SecurityProfiles,
		// HSS also seems to need these, or it will ignore the JSON
		ApnProfiles: src.ApnProfiles,
		QosProfiles: src.QosProfiles,
	}
	return &dest
}

// FilterConfigPCRF returns the portion of the JsonConfig relevant to PCRF
func filterConfigPCRF(src *jsonConfig) *jsonConfig {
	dest := jsonConfig{
		Policies: src.Policies,
	}
	return &dest
}

// PostData filters and posts the data to the service
func (s *Synchronizer) PostData(name string, endpoint string, filter FilterDataFunc, config *jsonConfig) error {
	var filteredConfig *jsonConfig
	if filter == nil {
		filteredConfig = config

	} else {
		filteredConfig = filter(config)
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

// SynchronizeConnectivityService synchronizes the connectivity service
func (s *Synchronizer) SynchronizeConnectivityService(device *models.Device, cs *models.ConnectivityService_ConnectivityService_ConnectivityService, validEnterpriseIds map[string]bool) error {
	config := jsonConfig{}

	log.Infof("Synchronizing Connectivity Service %s", *cs.Id)

	if device.Subscriber != nil {
		for _, ue := range device.Subscriber.Ue {
			keys := subscriberKeys{}

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
				keys.ServingPlmn = &subscriberServingPlmn{
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
				keys.ImsiRange = &subscriberImsiRange{
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

			rule := subscriberSelectionRule{
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

			config.SubscriberSelectionRules = append(config.SubscriberSelectionRules, rule)
		}
	}

	if device.ApnProfile != nil {
		config.ApnProfiles = make(map[string]apnProfile)
		for _, apn := range device.ApnProfile.ApnProfile {
			profile := apnProfile{
				ApnName:      apn.ApnName,
				DNSPrimary:   apn.DnsPrimary,
				DNSSecondary: apn.DnsSecondary,
				Mtu:          apn.Mtu,
				GxEnabled:    apn.GxEnabled,
				Network:      "lbo", // TODO: update modeling and revise
				Usage:        1,     // TODO: update modeling and revise
				GxApn:        apn.ServiceGroup,
			}

			config.ApnProfiles[*apn.Id] = profile
		}
	}

	if device.AccessProfile != nil {
		config.AccessProfiles = make(map[string]accessProfile)
		for _, access := range device.AccessProfile.AccessProfile {
			profile := accessProfile{
				Type:   access.Type,
				Filter: access.Filter,
			}

			config.AccessProfiles[*access.Id] = profile
		}
	}

	if device.QosProfile != nil {
		config.QosProfiles = make(map[string]qosProfile)
		for _, qos := range device.QosProfile.QosProfile {
			arp := qosArp{
				Priority:                0, // default
				PreemptionCapability:    1, // default
				PreemptionVulnerability: 1, // default
			}
			if qos.Arp != nil {
				if qos.Arp.Priority != nil {
					arp.Priority = *qos.Arp.Priority
				}
				if qos.Arp.PreemptionCapability != nil {
					arp.PreemptionCapability = synchronizer.BoolToUint32(*qos.Arp.PreemptionCapability)
				}
				if qos.Arp.PreemptionVulnerability != nil {
					arp.PreemptionVulnerability = synchronizer.BoolToUint32(*qos.Arp.PreemptionVulnerability)
				}
			}

			profile := qosProfile{
				Qci: 9, // default
				Arp: &arp,
			}

			if qos.ApnAmbr != nil {
				profile.ApnAmbr = []uint32{*qos.ApnAmbr.Downlink, *qos.ApnAmbr.Uplink}
			}

			if qos.Qci != nil {
				profile.Qci = *qos.Qci
			}

			config.QosProfiles[*qos.Id] = profile
		}
	}

	if device.UpProfile != nil {
		config.UpProfiles = make(map[string]upProfile)
		for _, up := range device.UpProfile.UpProfile {
			profile := upProfile{
				UserPlane:     up.UserPlane,
				AccessControl: up.AccessControl,
				AccessTags:    map[string]string{"tag1": "ACC"}, // TODO: update modeling and revise
				QosTags:       map[string]string{"tag1": "BW"},  // TODO: update modeling and revise
			}

			config.UpProfiles[*up.Id] = profile
		}
	}

	if device.SecurityProfile != nil {
		config.SecurityProfiles = make(map[string]securityProfile)
		for _, sp := range device.SecurityProfile.SecurityProfile {
			profile := securityProfile{
				Key: sp.Key,
				Opc: sp.Opc,
				Sqn: sp.Sqn,
			}

			config.SecurityProfiles[*sp.Id] = profile
		}
	}

	if (device.ServicePolicy != nil) || (device.ServiceRule != nil) || (device.ServiceGroup != nil) {
		var err error
		config.Policies, err = s.synchronizePCRF(device)
		if err != nil {
			return err
		}
	}

	synchronizer.KpiSynchronizationResourceTotal.WithLabelValues(*cs.Id, "subscriber").Set(float64(len(config.SubscriberSelectionRules)))
	synchronizer.KpiSynchronizationResourceTotal.WithLabelValues(*cs.Id, "apn-profile").Set(float64(len(config.ApnProfiles)))
	synchronizer.KpiSynchronizationResourceTotal.WithLabelValues(*cs.Id, "access-profile").Set(float64(len(config.AccessProfiles)))
	synchronizer.KpiSynchronizationResourceTotal.WithLabelValues(*cs.Id, "qos-profile").Set(float64(len(config.QosProfiles)))
	synchronizer.KpiSynchronizationResourceTotal.WithLabelValues(*cs.Id, "up-profile").Set(float64(len(config.UpProfiles)))
	synchronizer.KpiSynchronizationResourceTotal.WithLabelValues(*cs.Id, "security-profile").Set(float64(len(config.SecurityProfiles)))

	if s.outputFileName != "" {
		data, err := json.MarshalIndent(config, "", "  ")
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
			err := s.PostData("SPGWC", *cs.SpgwcEndpoint, filterConfigSPGWC, &config)
			if err != nil {
				return err
			}
		}

		if cs.HssEndpoint != nil {
			err := s.PostData("HSS", *cs.HssEndpoint, filterConfigHSS, &config)
			if err != nil {
				return err
			}
		}

		if cs.PcrfEndpoint != nil {
			err := s.PostData("PCRF", *cs.PcrfEndpoint, filterConfigPCRF, &config)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
