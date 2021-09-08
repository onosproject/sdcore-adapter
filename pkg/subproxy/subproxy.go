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
	clientHTTP = &http.Client{}
}

//HTTPClient interface
//go:generate mockgen -destination=../test/mocks/mock_http.go -package=mocks github.com/onosproject/sdcore-adapter/pkg/subproxy HTTPClient
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func (s *subscriberProxy) addSubscriberByID(c *gin.Context) {
	log.Debugf("Received One Subscriber Data")
	ueID := c.Param("ueId")
	var payload []byte
	if c.Request.Body != nil {
		payload, _ = ioutil.ReadAll(c.Request.Body)
	}

	if !strings.HasPrefix(ueID, "imsi-") {
		log.Debugf("Ue Id format is invalid ")
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	log.Debugf("Received subscriber id : %s ", ueID)

	split := strings.Split(ueID, "-")
	imsiValue, err := strconv.ParseUint(split[1], 10, 64)
	if err != nil {
		log.Error(err.Error())
		return
	}
	err = s.updateImsiDeviceGroup(&imsiValue)
	if err != nil {
		jsonByte, okay := getJSONResponse(err.Error())
		if okay != nil {
			log.Debug(err.Error())
		}
		c.Data(http.StatusInternalServerError, "application/json", jsonByte)
		return
	}

	resp, err := postToWebConsole(s.BaseWebConsoleURL+subscriberAPISuffix+ueID, payload, s.PostTimeout)
	if err != nil {
		jsonByte, okay := getJSONResponse(err.Error())
		if okay != nil {
			log.Debug(err.Error())
		}
		c.Data(resp.StatusCode, "application/json", jsonByte)
		return
	}
	if resp.StatusCode != 201 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			jsonByte, okay := getJSONResponse(err.Error())
			if okay != nil {
				log.Debug(err.Error())
			}
			c.Data(http.StatusInternalServerError, "application/json", jsonByte)
			return
		}

		bodyBytes, err = getJSONResponse(string(bodyBytes))
		if err != nil {
			log.Debug(err.Error())
		}
		c.Data(resp.StatusCode, "application/json", bodyBytes)
		return
	}

	c.JSON(resp.StatusCode, gin.H{"status": "success"})
}

func (s *subscriberProxy) updateImsiDeviceGroup(imsi *uint64) error {

	// Getting the current configuration from the ROC
	origVal, err := s.gnmiClient.GetPath(context.Background(), "", s.AetherConfigTarget, s.AetherConfigAddress)
	if err != nil {
		return errors.NewInvalid("failed to get the current state from onos-config: %v", err)
	}

	// Convert the JSON config into a Device structure
	origJSONBytes := origVal.GetJsonVal()
	device := &models.Device{}
	if len(origJSONBytes) > 0 {
		if err := models.Unmarshal(origJSONBytes, device); err != nil {
			log.Error("Failed to unmarshal json")
			return errors.NewInvalid("failed to unmarshal json")
		}
	}

	// Check if the IMSI already exists
	dg := findImsiInDeviceGroup(device, *imsi)
	if dg != nil {
		log.Debugf("Imsi %v already exists in device group %s", *imsi, *dg.Id)
		return nil
	}
	log.Debugf("Imsi doesn't exist in any device group")

	//Check if the site exists
	site := findSiteForTheImsi(device, *imsi)

	if site == nil {
		log.Debugf("Not site found for this imsi %s", *imsi)
		dgroup := "defaultent-defaultsite-default"
		return s.addImsiToDefaultGroup(device, dgroup, imsi)
	}
	dgroup := *site.Id + "-default"
	return s.addImsiToDefaultGroup(device, dgroup, imsi)

}

//addImsiToDefaultGroup adds Imsi to default group expect the group already exists
func (s *subscriberProxy) addImsiToDefaultGroup(device *models.Device, dgroup string, imsi *uint64) error {
	log.Debugf("AddImsiToDefaultGroup Name : %s", dgroup)

	// Now get the device group the caller wants us to add the IMSI to
	dg, okay := device.DeviceGroup.DeviceGroup[dgroup]
	if !okay {
		return errors.NewInvalid("failed to find device group %v", dgroup)
	}
	site, err := getSiteForDeviceGrp(device, dg)
	if err != nil {
		log.Error("Failed to find site for device group %v", *dg.Id)
		return errors.NewInternal("failed to find site for device group %v", *dg.Id)
	}
	maskedImsi, err := sync.MaskSubscriberImsiDef(site.ImsiDefinition, *imsi) // mask off the MCC/MNC/EntId
	if err != nil {
		return errors.NewInvalid("Failed to mask the subscriber: %v", err)
	}

	log.Debugf("Masked imsi is %v", maskedImsi)

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
		return errors.NewInternal("Error executing gNMI: %v", err)
	}
	return nil

}

func (s *subscriberProxy) StartSubscriberProxy(bindPort string, path string) error {
	router := gin.New()
	router.Use(getlogger(), gin.Recovery())
	router.POST(path, getlogger(), s.addSubscriberByID)
	err := router.Run("0.0.0.0" + bindPort)
	if err != nil {
		return err
	}
	return nil
}

//NewSubscriberProxy as Init method
func NewSubscriberProxy(aetherConfigTarget string, baseWebConsoleURL string, aetherConfigAddr string,
	gnmiClient gnmiclient.GnmiInterface, postTimeout time.Duration) *subscriberProxy {
	sproxy := &subscriberProxy{
		AetherConfigAddress: aetherConfigAddr,
		AetherConfigTarget:  aetherConfigTarget,
		BaseWebConsoleURL:   baseWebConsoleURL,
		gnmiClient:          gnmiClient,
		PostTimeout:         postTimeout,
	}
	return sproxy
}