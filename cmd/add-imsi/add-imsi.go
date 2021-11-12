// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	"os"

	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	sync "github.com/onosproject/sdcore-adapter/pkg/synchronizer"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

// Syntax:
// # Note: do not use leading zero on imsi or it will be interpreted as octal!!
// add-imsi -client_key=/etc/sdcore-adapter/certs/tls.key -client_crt=/etc/sdcore-adapter/certs/tls.crt -ca_crt=/etc/sdcore-adapter/certs/tls.cacert -hostCheckDisabled -aether-config onos-config:5150 -target connectivity-service-v3 -devicegroup starbucks-newyork-default -imsi 21032002000123

var (
	dgroup           = flag.String("devicegroup", "", "device group to add the imsi to")
	imsi             = flag.Uint64("imsi", 0, "imsi to add")
	target           = flag.String("target", "", "target device name")
	aetherConfigAddr = flag.String("aether-config", "", "address of aether-config")
)

var log = logging.GetLogger("add-imsi")

func getDeviceGroupSite(device *models.Device, dg *models.OnfDeviceGroup_DeviceGroup_DeviceGroup) (*models.OnfSite_Site_Site, error) {
	if (dg.Site == nil) || (*dg.Site == "") {
		return nil, fmt.Errorf("DeviceGroup %s has no site", *dg.Id)
	}
	site, okay := device.Site.Site[*dg.Site]
	if !okay {
		return nil, fmt.Errorf("DeviceGroup %s site %s not found", *dg.Id, *dg.Site)
	}
	if (site.Enterprise == nil) || (*site.Enterprise == "") {
		return nil, fmt.Errorf("DeviceGroup %s has no enterprise", *dg.Id)
	}
	return site, nil
}

func findImsiInDeviceGroup(device *models.Device, imsi uint64) *models.OnfDeviceGroup_DeviceGroup_DeviceGroup {
deviceGroupLoop:
	for _, dg := range device.DeviceGroup.DeviceGroup {
		for _, imsiBlock := range dg.Imsis {
			site, err := getDeviceGroupSite(device, dg)
			if err != nil {
				log.Warnf("Error getting site: %v", err)
				continue deviceGroupLoop
			}

			if imsiBlock.ImsiRangeFrom == nil {
				log.Infof("imsiBlock %s in dg %s has blank ImsiRangeFrom", *imsiBlock.ImsiId, *dg.Id)
				continue deviceGroupLoop
			}
			var firstImsi uint64
			firstImsi, err = sync.FormatImsiDef(site.ImsiDefinition, *imsiBlock.ImsiRangeFrom)
			if err != nil {
				log.Infof("Failed to format IMSI in dg %s: %v", *dg.Id, err)
				continue deviceGroupLoop
			}
			var lastImsi uint64
			if imsiBlock.ImsiRangeTo == nil {
				lastImsi = firstImsi
			} else {
				lastImsi, err = sync.FormatImsiDef(site.ImsiDefinition, *imsiBlock.ImsiRangeTo)
				if err != nil {
					log.Infof("Failed to format IMSI in dg %s: %v", *dg.Id, err)
					continue deviceGroupLoop
				}

			}
			log.Infof("Compare %v %v %v", imsi, firstImsi, lastImsi)
			if (imsi >= firstImsi) && (imsi <= lastImsi) {
				return dg
			}
		}
	}

	return nil
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if (imsi == nil) || (dgroup == nil) || (target == nil) || (aetherConfigAddr == nil) {
		log.Fatal("Use the -imsi, -dgroup, -target, and -aetherConfigAddr options")
	}

	// Get the current configuration from the ROC
	origVal, err := gnmiclient.GetPath(context.Background(), "", *target, *aetherConfigAddr)
	if err != nil {
		log.Fatal("Failed to get the current state from onos-config: %v", err)
	}

	// Convert the JSON config into a Device structure
	origJSONBytes := origVal.GetJsonVal()
	device := &models.Device{}
	if len(origJSONBytes) > 0 {
		if err := models.Unmarshal(origJSONBytes, device); err != nil {
			log.Fatal("Failed to unmarshal json")
		}
	}

	// See if the IMSI already exists
	dg := findImsiInDeviceGroup(device, *imsi)
	if dg != nil {
		log.Infof("Imsi %v already exists in device group %s", *imsi, *dg.Id)
		return
	}

	// Now get the device group the caller wants us to add the IMSI to
	dg, okay := device.DeviceGroup.DeviceGroup[*dgroup]
	if !okay {
		log.Fatal("Failed to find device group %v", *dgroup)
	}
	site, err := getDeviceGroupSite(device, dg) // and the site
	if err != nil {
		log.Fatal("Failed to find site for device group %v", *dg.Id)
	}
	maskedImsi, err := sync.MaskSubscriberImsiDef(site.ImsiDefinition, *imsi) // mask off the MCC/MNC/EntId
	if err != nil {
		log.Fatal("Failed to mask the subscriber: %v", err)
	}

	log.Infof("Masked imsi is %v", maskedImsi)

	// An imsi-range inside of a devicegroup needs a name. Let's just name our range after the imsi
	// we're creating, prepended with "auto-" to make it clear it was automatically added. Don't worry
	// about coalescing ranges -- just create simple ranges with exactly one imsi per range.
	rangeName := fmt.Sprintf("auto-%d", *imsi)

	// Generate a prefix into the gNMI configuration tree
	prefix := gnmiclient.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]/imsis[name=%s]", *dgroup, rangeName), *target)

	// Build up a list of gNMI updates to apply
	updates := []*gpb.Update{}
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("imsi-range-from", *target, &maskedImsi))

	// Apply them
	err = gnmiclient.Update(context.Background(), prefix, *target, *aetherConfigAddr, updates)
	if err != nil {
		log.Fatalf("Error executing gNMI: %v", err)
	}

	log.Infof("It worked!")
}
