// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package imsirange

import (
	"context"
	"fmt"
	"sort"
	"strings"

	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"

	sync "github.com/onosproject/sdcore-adapter/pkg/synchronizer/v3"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
)

// GetDevice will getting the existing values
func (irange ImsiRange) GetDevice(gnmiClient gnmiclient.GnmiInterface) (*models.Device, error) {
	origVal, err := gnmiClient.GetPath(context.Background(), "", irange.AetherConfigTarget, irange.AetherConfigAddress)
	if err != nil {
		return nil, errors.NewInvalid("Failed to get the current state from onos-config: %v", err)
	}

	origJSONBytes := origVal.GetJsonVal()
	device := &models.Device{}
	if len(origJSONBytes) > 0 {
		if err := models.Unmarshal(origJSONBytes, device); err != nil {
			return nil, errors.NewInvalid("Failed to unmarshal json")
		}
	}

	return device, nil

}

// CollapseImsi will iterate through all the device group and check if ranges can be collapse or not
func (irange ImsiRange) CollapseImsi(device *models.Device, gnmiClient gnmiclient.GnmiInterface) error {
deviceGroupLoop:
	for _, dg := range device.DeviceGroup.DeviceGroup {
		//Only considering the default device group
		if strings.HasSuffix(*dg.Id, "-default") {
			imsies := []Imsi{}
			var im Imsi
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

			newRanges, delRange := mergeImsi(imsies)

			for val := range delRange {
				for j := range imsies {
					if val == imsies[j].name {
						log.Infof("Deleting %v", imsies[j])
						err := irange.DeleteImsiRanges(device, gnmiClient, *dg.Id, imsies[j])
						if err != nil {
							log.Infof("Error while deleting range error:%v", err)
							return errors.NewInvalid("Error while deleting range error:%v", err)
						}
					}
				}
			}

		nextRange:
			for i := range newRanges {
				for j := range imsies {
					if newRanges[i] == imsies[j] {
						continue nextRange
					}
				}
				err := irange.AddImsiRange(device, gnmiClient, *dg.Id, newRanges[i].firstImsi, newRanges[i].lastImsi)
				if err != nil {
					log.Errorf(err.Error())
					return err
				}
				log.Infof("Adding %v", newRanges[i])
			}

		}
	}
	return nil
}

// AddImsiRange it will add the imsi range
func (irange ImsiRange) AddImsiRange(device *models.Device, gnmiClient gnmiclient.GnmiInterface, dgroup string,
	firstimsi uint64, lastimsi uint64) error {
	log.Infof("AddImsiToDefaultGroup Name : %s", dgroup)

	// Now get the device group the caller wants us to add the IMSI to
	dg, okay := device.DeviceGroup.DeviceGroup[dgroup]
	if !okay {
		log.Errorf("Failed to find device group %v", dgroup)
		return errors.NewInvalid("failed to find device group %v", dgroup)
	}
	site, err := getDeviceGroupSite(device, dg)
	if err != nil {
		return errors.NewInvalid("failed to find site for device group %v", *dg.Id)
	}

	maskedFirstImsi, maskedLastImsi, err := getMaskedImsies(site.ImsiDefinition, firstimsi, lastimsi)
	if err != nil {
		return err
	}

	log.Infof("Masked First imsi is %v", maskedFirstImsi)
	log.Infof("Masked Last imsi is %v", maskedLastImsi)

	rangeName := generateName(site.ImsiDefinition, firstimsi, maskedFirstImsi)

	// Generate a prefix into the gNMI configuration tree
	prefix := gnmiclient.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]/imsis[name=%s]",
		dgroup, rangeName), irange.AetherConfigTarget)

	// Build up a list of gNMI updates to apply
	updates := []*gpb.Update{}
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("imsi-range-from", irange.AetherConfigTarget,
		&maskedFirstImsi))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("imsi-range-to", irange.AetherConfigTarget,
		&maskedLastImsi))

	// Apply them
	err = gnmiClient.Update(context.Background(), prefix, irange.AetherConfigTarget, irange.AetherConfigAddress, updates)
	if err != nil {
		log.Errorf("Error executing gNMI: %v", err)
		return fmt.Errorf("error executing gNMI: %v", err)
	}
	return nil
}

// MergeImsi it will merge the imsi ranges
func mergeImsi(imsies []Imsi) ([]Imsi, map[string]struct{}) {
	//Assumption: Imsi ranges will never overlap
	var newImsi []Imsi
	delImsi := make(map[string]struct{})
	var zeroByte = struct{}{}

	for i := 0; i < len(imsies); i++ {

		if len(newImsi) == 0 {
			newImsi = append(newImsi, imsies[i])
		} else if (imsies[i].firstImsi - newImsi[len(newImsi)-1].lastImsi) == 1 {
			newImsi[len(newImsi)-1].lastImsi = imsies[i].lastImsi
			delImsi[newImsi[len(newImsi)-1].name] = zeroByte
			delImsi[imsies[i].name] = zeroByte
		} else {
			newImsi = append(newImsi, imsies[i])
		}

	}
	return newImsi, delImsi
}

// DeleteImsiRanges it will delete the imsi range
func (irange ImsiRange) DeleteImsiRanges(device *models.Device, gnmiClient gnmiclient.GnmiInterface, dgroup string, imsi Imsi) error {

	prefix := gnmiclient.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]/imsis[name=%s]", dgroup, imsi.name),
		irange.AetherConfigTarget)
	dg, okay := device.DeviceGroup.DeviceGroup[dgroup]
	if !okay {
		return errors.NewInvalid("failed to find device group %v", dgroup)
	}
	site, err := getDeviceGroupSite(device, dg)
	if err != nil {
		return errors.NewInvalid("failed to find site for device group %v", *dg.Id)
	}

	maskedFirstImsi, maskedLastImsi, err := getMaskedImsies(site.ImsiDefinition, imsi.firstImsi, imsi.lastImsi)
	if err != nil {
		return err
	}

	updates := []*gpb.Update{}
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("imsi-range-from", irange.AetherConfigTarget, &maskedFirstImsi))
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("imsi-range-to", irange.AetherConfigTarget, &maskedLastImsi))
	deletes := gnmiclient.DeleteFromUpdates(updates, irange.AetherConfigTarget)

	err = gnmiClient.Delete(context.Background(), prefix, irange.AetherConfigTarget, irange.AetherConfigAddress, deletes)
	if err != nil {
		return errors.NewInvalid("error executing gNMI: %v", err)
	}

	return nil
}

func getDeviceGroupSite(device *models.Device, dg *models.DeviceGroup_DeviceGroup_DeviceGroup) (*models.Site_Site_Site, error) {
	if (dg.Site == nil) || (*dg.Site == "") {
		return nil, errors.NewInvalid("DeviceGroup %s has no site", *dg.Id)
	}
	site, okay := device.Site.Site[*dg.Site]
	if !okay {
		return nil, errors.NewInvalid("DeviceGroup %s site %s not found", *dg.Id, *dg.Site)
	}
	if (site.Enterprise == nil) || (*site.Enterprise == "") {
		return nil, errors.NewInvalid("DeviceGroup %s has no enterprise", *dg.Id)
	}
	return site, nil
}

func getMaskedImsies(i *models.Site_Site_Site_ImsiDefinition, firstImsi uint64, lastImsi uint64) (uint64, uint64, error) {

	var maskedFirstImsi uint64
	var maskedLastImsi uint64

	maskedFirstImsi, err := sync.MaskSubscriberImsiDef(i, firstImsi) // mask off the MCC/MNC/EntId
	if err != nil {
		return maskedFirstImsi, maskedLastImsi, errors.NewInvalid("failed to mask the subscriber: %v", err)
	}

	maskedLastImsi, err = sync.MaskSubscriberImsiDef(i, lastImsi) // mask off the MCC/MNC/EntId
	if err != nil {
		log.Errorf("Failed to mask the subscriber: %v", err)
		return maskedFirstImsi, maskedLastImsi, errors.NewInvalid("failed to mask the subscriber: %v", err)
	}

	return maskedFirstImsi, maskedLastImsi, nil

}

func generateName(s *models.Site_Site_Site_ImsiDefinition, firstImsi uint64, maskedFirstImsi uint64) string {

	if s.Mcc == nil {
		rangeName := fmt.Sprintf("auto-%d", firstImsi)
		return rangeName
	}
	rangeName := fmt.Sprintf("auto-%s%s%d-%d", *s.Mcc, *s.Mnc, *s.Enterprise, maskedFirstImsi)
	return rangeName
}

// NewIMSIRange ...
func NewIMSIRange(aetherConfigAddr string, aetherConfigTarget string) *ImsiRange {
	i := &ImsiRange{
		AetherConfigAddress: aetherConfigAddr,
		AetherConfigTarget:  aetherConfigTarget,
	}
	return i
}
