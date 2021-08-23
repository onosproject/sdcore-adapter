// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0
package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strings"
	"time"

	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	sync "github.com/onosproject/sdcore-adapter/pkg/synchronizer/v3"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"github.com/prometheus/common/log"
)

type imsi struct {
	firstImsi uint64
	lastImsi  uint64
	name      string
}

var (
	aetherConfigTarget = flag.String("aether_config_target", "connectivity-service-v3", "Target to use when pulling from aether-config")
	aetherConfigAddr   = flag.String("onos_config_url", "onos-config.micro-onos.svc.cluster.local:5150", "url of onos-config")
)

//It will merge the imsi ranges if possible and delete the old renges
func collapseImsi() {

	//Getting the existing values
	origVal, err := migration.GetPath("", *aetherConfigTarget, *aetherConfigAddr, context.Background())
	if err != nil {
		fmt.Printf("Failed to get the current state from onos-config: %v", err)
	}

	origJsonBytes := origVal.GetJsonVal()
	device := &models.Device{}
	if len(origJsonBytes) > 0 {
		if err := models.Unmarshal(origJsonBytes, device); err != nil {
			fmt.Printf("Failed to unmarshal json")
			return
		}
	}
deviceGroupLoop:
	for _, dg := range device.DeviceGroup.DeviceGroup {
		//Only considering the default device group
		if strings.HasSuffix(*dg.Id, "-default") {
			imsies := []imsi{}
			var im imsi
			for _, imsiBlock := range dg.Imsis {

				site, err := getDeviceGroupSite(device, dg)
				if err != nil {
					log.Warnf("Error getting site: %v", err)
					continue deviceGroupLoop
				}
				var origfirstImsi uint64
				origfirstImsi, err = sync.FormatImsiDef(site.ImsiDefinition, *imsiBlock.ImsiRangeFrom)
				if err != nil {
					log.Infof("Failed to format IMSI in dg %s: %v", *dg.Id, err)
					continue deviceGroupLoop
				}
				var origlastImsi uint64
				if imsiBlock.ImsiRangeTo == nil {
					origlastImsi = origfirstImsi
				} else {
					origlastImsi, err = sync.FormatImsiDef(site.ImsiDefinition, *imsiBlock.ImsiRangeTo)
					if err != nil {
						log.Infof("Failed to format IMSI in dg %s: %v", *dg.Id, err)
						continue deviceGroupLoop
					}

				}

				im.firstImsi = origfirstImsi
				im.lastImsi = origlastImsi
				im.name = *imsiBlock.Name
				imsies = append(imsies, im)
			}

			//sorting Imsies (if needed move to new function)
			sort.SliceStable(imsies, func(i, j int) bool {
				return imsies[i].firstImsi < imsies[j].firstImsi
			})

			newRanges, delRenge := mergeImsi(imsies)

		nextRange:
			for i := range newRanges {
				for j := range imsies {
					if newRanges[i] == imsies[j] {
						continue nextRange
					}
				}
				//TODO: If possible use go routine here
				err := AddImsiRange(device, *dg.Id, newRanges[i].firstImsi, newRanges[i].lastImsi)
				if err != nil {
					log.Errorf(err.Error())
					return
				}
			}

			for i := range delRenge {
				for j := range imsies {
					if delRenge[i] == imsies[j].name {
						log.Infof("Deleting %v", imsies[j].name)
						err := DeleteImsiRanges(device, *dg.Id, imsies[j])
						if err != nil {
							log.Infof("Error while deleting range error:%v", err)
						}
					}
				}
			}

		}
	}

}

func mergeImsi(imsies []imsi) ([]imsi, []string) {
	//Assumption: Imsi ranges will never overlap
	var newImsi []imsi
	var delImsi = []string{}

	for i := 0; i < len(imsies); i++ {

		if len(newImsi) == 0 {
			newImsi = append(newImsi, imsies[i])
		} else if (imsies[i].firstImsi - newImsi[len(newImsi)-1].lastImsi) == 1 {
			newImsi[len(newImsi)-1].lastImsi = imsies[i].lastImsi
			delImsi = append(delImsi, newImsi[len(newImsi)-1].name)
			delImsi = append(delImsi, imsies[i].name)
		} else {
			newImsi = append(newImsi, imsies[i])
		}

	}
	return newImsi, delImsi
}

func getDeviceGroupSite(device *models.Device, dg *models.DeviceGroup_DeviceGroup_DeviceGroup) (*models.Site_Site_Site, error) {
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

func AddImsiRange(device *models.Device, dgroup string, firstimsi uint64, lastimsi uint64) error {
	log.Infof("AddImsiToDefaultGroup Name : %s", dgroup)

	// Now get the device group the caller wants us to add the IMSI to
	dg, okay := device.DeviceGroup.DeviceGroup[dgroup]
	if !okay {
		log.Errorf("Failed to find device group %v", dgroup)
		return fmt.Errorf("failed to find device group %v", dgroup)
	}
	site, err := getDeviceGroupSite(device, dg)
	if err != nil {
		log.Errorf("Failed to find site for device group %v", *dg.Id)
		return fmt.Errorf("failed to find site for device group %v", *dg.Id)
	}

	maskedFirstImsi, maskedLastImsi, err := getMaskedImsies(site.ImsiDefinition, firstimsi, lastimsi)
	if err != nil {
		return err
	}

	log.Infof("Masked First imsi is %v", maskedFirstImsi)
	log.Infof("Masked Last imsi is %v", maskedLastImsi)

	// An imsi-range inside of a devicegroup needs a name. Generating it based on plmn+imsis_count+randomstr
	rangeName := generateName(site.ImsiDefinition, firstimsi, lastimsi)

	// Generate a prefix into the gNMI configuration tree
	prefix := migration.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]/imsis[name=%s]", dgroup, rangeName), *aetherConfigTarget)

	// Build up a list of gNMI updates to apply
	updates := []*gpb.Update{}
	updates = migration.AddUpdate(updates, migration.UpdateUInt64("imsi-range-from", *aetherConfigTarget, &maskedFirstImsi))
	updates = migration.AddUpdate(updates, migration.UpdateUInt64("imsi-range-to", *aetherConfigTarget, &maskedLastImsi))

	// Apply them
	err = migration.Update(prefix, *aetherConfigTarget, *aetherConfigAddr, updates, context.Background())
	if err != nil {
		log.Errorf("Error executing gNMI: %v", err)
		return fmt.Errorf("error executing gNMI: %v", err)
	}
	return nil
}

func generateName(s *models.Site_Site_Site_ImsiDefinition, firstImsi uint64, lastImsi uint64) string {

	//TODO: Need to disscuss and if required change the naming logic
	if s.Mcc == nil {
		rangeName := fmt.Sprintf("auto-%d-%d-%s", firstImsi, lastImsi-firstImsi, genRandomstr(6))
		return rangeName
	}
	rangeName := fmt.Sprintf("auto-%d%d%d-%d-%s", *s.Mcc, *s.Mnc, *s.Enterprise, lastImsi-firstImsi, genRandomstr(6))
	return rangeName
}

func genRandomstr(num int) string {
	seedVal := rand.NewSource(time.Now().UnixNano()) //generating seed value based on time
	ran := rand.New(seedVal)                         //Here we are providing different seed value.
	var letters = []rune("abcdefghijkl0123456789mnopqrstuvwxyz")
	str := make([]rune, num)
	for i := range str {
		str[i] = letters[ran.Intn(len(letters))]
	}

	return string(str)

}

func DeleteImsiRanges(device *models.Device, dgroup string, imsi imsi) error {

	prefix := migration.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]/imsis[name=%s]", dgroup, imsi.name), *aetherConfigTarget)
	dg, okay := device.DeviceGroup.DeviceGroup[dgroup]
	if !okay {
		log.Errorf("Failed to find device group %v", dgroup)
		return fmt.Errorf("failed to find device group %v", dgroup)
	}
	site, err := getDeviceGroupSite(device, dg)
	if err != nil {
		log.Errorf("Failed to find site for device group %v", *dg.Id)
		return fmt.Errorf("failed to find site for device group %v", *dg.Id)
	}

	maskedFirstImsi, maskedLastImsi, err := getMaskedImsies(site.ImsiDefinition, imsi.firstImsi, imsi.lastImsi)
	if err != nil {
		return err
	}

	updates := []*gpb.Update{}
	updates = migration.AddUpdate(updates, migration.UpdateUInt64("imsi-range-from", *aetherConfigTarget, &maskedFirstImsi))
	updates = migration.AddUpdate(updates, migration.UpdateUInt64("imsi-range-to", *aetherConfigTarget, &maskedLastImsi))
	deletes := migration.DeleteFromUpdates(updates, *aetherConfigTarget)

	err = migration.Delete(prefix, *aetherConfigTarget, *aetherConfigAddr, deletes, context.Background())
	if err != nil {
		log.Errorf("Error executing gNMI: %v", err)
		return fmt.Errorf("error executing gNMI: %v", err)
	}

	return nil
}

func getMaskedImsies(i *models.Site_Site_Site_ImsiDefinition, firstImsi uint64, lastImsi uint64) (uint64, uint64, error) {

	var maskedFirstImsi uint64
	var maskedLastImsi uint64

	maskedFirstImsi, err := sync.MaskSubscriberImsiDef(i, firstImsi) // mask off the MCC/MNC/EntId
	if err != nil {
		log.Errorf("Failed to mask the subscriber: %v", err)
		return maskedFirstImsi, maskedLastImsi, fmt.Errorf("failed to mask the subscriber: %v", err)
	}

	maskedLastImsi, err = sync.MaskSubscriberImsiDef(i, lastImsi) // mask off the MCC/MNC/EntId
	if err != nil {
		log.Errorf("Failed to mask the subscriber: %v", err)
		return maskedFirstImsi, maskedLastImsi, fmt.Errorf("failed to mask the subscriber: %v", err)
	}

	return maskedFirstImsi, maskedLastImsi, nil

}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	collapseImsi()
}
