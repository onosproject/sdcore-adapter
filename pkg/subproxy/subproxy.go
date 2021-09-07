// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package subproxy

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	sync "github.com/onosproject/sdcore-adapter/pkg/synchronizer/v3"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func init() {
	Client = &http.Client{}
}

// AddSubscriberByID IMSI(ueId)
func (s *SubscriberProxy) AddSubscriberByID(c *gin.Context) {
	log.Info("Received One Subscriber Data")
	ueID := c.Param("ueId")
	var payload []byte
	if c.Request.Body != nil {
		payload, _ = ioutil.ReadAll(c.Request.Body)
	}

	if !strings.HasPrefix(ueID, "imsi-") {
		log.Error("Ue Id format is invalid ")
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	log.Infof("Received subscriber id : %s ", ueID)

	split := strings.Split(ueID, "-")
	imsiValue, err := strconv.ParseUint(split[1], 10, 64)
	if err != nil {
		log.Error(err.Error())
		return
	}
	err = s.UpdateImsiDeviceGroup(&imsiValue)
	if err != nil {
		c.Data(http.StatusInternalServerError, "application/json", getJSONResponse(err.Error()))
		return
	}

	resp, err := PostToWebConsole(s.BaseWebConsoleURL+subscriberAPISuffix+ueID, payload, s.PostTimeout)
	if err != nil {
		c.Data(resp.StatusCode, "application/json", getJSONResponse(err.Error()))
		return
	}
	if resp.StatusCode != 201 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err.Error())
			c.Data(http.StatusInternalServerError, "application/json", getJSONResponse(err.Error()))
			return
		}
		c.Data(resp.StatusCode, "application/json", getJSONResponse(string(bodyBytes)))
		return
	}

	c.JSON(resp.StatusCode, gin.H{"status": "success"})
}

// UpdateImsiDeviceGroup is ...
func (s *SubscriberProxy) UpdateImsiDeviceGroup(imsi *uint64) error {
	log.Info("Calling UpdateImsiDeviceGroup...")

	// Get the current configuration from the ROC

	origVal, err := s.gnmiClient.GetPath(context.Background(), "", s.AetherConfigTarget, s.AetherConfigAddress)
	if err != nil {
		log.Error("Failed to get the current state from onos-config: %v", err)
		return errors.NewInvalid("failed to get the current state from onos-config: %v", err)
	}

	// Convert the JSON config into a Device structure
	origJSONBytes := origVal.GetJsonVal()
	device := &models.Device{}
	if len(origJSONBytes) > 0 {
		if err := models.Unmarshal(origJSONBytes, device); err != nil {
			log.Error("Failed to unmarshal json")
			return fmt.Errorf("failed to unmarshal json")
		}
	}

	// Check if the IMSI already exists
	dg := findImsiInDeviceGroup(device, *imsi)
	if dg != nil {
		log.Infof("Imsi %v already exists in device group %s", *imsi, *dg.Id)
		return nil
	}
	log.Info("Imsi doesn't exist in any device group")

	//Check if the site exists
	site := findSiteForTheImsi(device, *imsi)

	if site == nil {
		log.Info("Not site found for this imsi %s", *imsi)
		dgroup := "defaultent-defaultsite-default"
		return s.AddImsiToDefaultGroup(device, dgroup, imsi)
	}
	dgroup := *site.Id + "-default"
	return s.AddImsiToDefaultGroup(device, dgroup, imsi)

}

//AddImsiToDefaultGroup adds Imsi to default group expect the group already exists
func (s *SubscriberProxy) AddImsiToDefaultGroup(device *models.Device, dgroup string, imsi *uint64) error {
	log.Infof("AddImsiToDefaultGroup Name : %s", dgroup)

	// Now get the device group the caller wants us to add the IMSI to
	dg, okay := device.DeviceGroup.DeviceGroup[dgroup]
	if !okay {
		log.Error("Failed to find device group %v", dgroup)
		return fmt.Errorf("failed to find device group %v", dgroup)
	}
	site, err := getSiteForDeviceGrp(device, dg)
	if err != nil {
		log.Error("Failed to find site for device group %v", *dg.Id)
		return errors.NewInternal("failed to find site for device group %v", *dg.Id)
	}
	maskedImsi, err := sync.MaskSubscriberImsiDef(site.ImsiDefinition, *imsi) // mask off the MCC/MNC/EntId
	if err != nil {
		log.Error("Failed to mask the subscriber: %v", err)
		return errors.NewInvalid("Failed to mask the subscriber: %v", err)
	}

	log.Infof("Masked imsi is %v", maskedImsi)

	// An imsi-range inside of a devicegroup needs a name. Let's just name our range after the imsi
	// we're creating, prepended with "auto-" to make it clear it was automatically added. Don't worry
	// about coalescing ranges -- just create simple ranges with exactly one imsi per range.
	rangeName := fmt.Sprintf("auto-%d", *imsi)

	// Generate a prefix into the gNMI configuration tree
	prefix := gnmiclient.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]/imsis[name=%s]", dgroup,
		rangeName), s.AetherConfigTarget)

	// Build up a list of gNMI updates to apply
	updates := []*gpb.Update{}
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("imsi-range-from", s.AetherConfigTarget, &maskedImsi))

	// Apply them
	err = s.gnmiClient.Update(context.Background(), prefix, s.AetherConfigTarget, s.AetherConfigAddress, updates)
	if err != nil {
		log.Errorf("Error executing gNMI: %v", err)
		return errors.NewInternal("Error executing gNMI: %v", err)
	}
	return nil

}

// StartSubscriberProxy starts server
func (s *SubscriberProxy) StartSubscriberProxy(bindPort string, path string) {
	router := gin.New()
	router.Use(getlogger(), gin.Recovery())
	router.POST(path, getlogger(), s.AddSubscriberByID)
	err := router.Run("0.0.0.0" + bindPort)
	if err != nil {
		log.Error(err.Error())
		return
	}
}

//NewSubscriberProxy as Init method
func NewSubscriberProxy(aetherConfigTarget string, baseWebConsoleURL string, aetherConfigAddr string,
	gnmiClient gnmiclient.GnmiInterface, postTimeout time.Duration) *SubscriberProxy {
	m := &SubscriberProxy{
		AetherConfigAddress: aetherConfigAddr,
		AetherConfigTarget:  aetherConfigTarget,
		BaseWebConsoleURL:   baseWebConsoleURL,
		gnmiClient:          gnmiClient,
		PostTimeout:         postTimeout,
	}
	return m
}
