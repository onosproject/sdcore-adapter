// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	sync "github.com/onosproject/sdcore-adapter/pkg/synchronizer/v3"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	bindPort           = flag.String("bind_port", ":5001", "Bind to just :port")
	postTimeout        = flag.Duration("post_timeout", time.Second*10, "Timeout duration when making post requests")
	aetherConfigTarget = flag.String("aether_config_target", "connectivity-service-v3", "Target to use when pulling from aether-config")

	baseWebConsoleUrl = flag.String("webconsole_url", "http://webui.omec.svc.cluster.local:5000", "base url for webui service address")
	aetherConfigAddr  = flag.String("onos_config_url", "onos-config.micro-onos.svc.cluster.local:5150", "url of onos-config")
)

type response struct {
	Status string `json:"status"`
}

var log = logging.GetLogger("subscriber-proxy")
var SubscriberAPISuffix = "/api/subscriber/"

// Add subscriber by IMSI(ueId)
func AddSubscriberByID(c *gin.Context) {

	log.Info("Received One Subscriber Data")
	ueId := c.Param("ueId")
	var payload []byte
	if c.Request.Body != nil {
		payload, _ = ioutil.ReadAll(c.Request.Body)
	}

	if !strings.HasPrefix(ueId, "imsi-") {
		log.Error("Ue Id format is invalid ")
		c.JSON(http.StatusInternalServerError, gin.H{})
		return
	}

	log.Infof("Received subscriber id : %s ", ueId)

	split := strings.Split(ueId, "-")
	imsiValue, err := strconv.ParseUint(split[1], 10, 64)
	if err != nil {
		log.Error(err.Error())
		return
	}
	err = UpdateImsiDeviceGroup(&imsiValue)
	if err != nil {
		c.Data(http.StatusInternalServerError, "application/json", getJsonResponse(err.Error()))
		return
	}

	err, resp := PostToWebConsole(ueId, payload)
	if err != nil {
		c.Data(resp.StatusCode, "application/json", getJsonResponse(err.Error()))
		return
	}

	if resp.StatusCode != 201 {

		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err.Error())
			c.Data(http.StatusInternalServerError, "application/json", getJsonResponse(err.Error()))
			return
		}
		c.Data(resp.StatusCode, "application/json", getJsonResponse(string(bodyBytes)))
		return
	}

	c.JSON(resp.StatusCode, gin.H{"status": "success"})
}

// Call webui API for subscriber provision on the SD-Core
func PostToWebConsole(imsi string, payload []byte) (error, *http.Response) {
	log.Info("Calling WebUI API...")
	client := &http.Client{
		Timeout: *postTimeout,
	}
	req, err := http.NewRequest("POST", *baseWebConsoleUrl+SubscriberAPISuffix+imsi, bytes.NewBuffer(payload))
	if err != nil {
		log.Info("Error while connecting webui ", err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return fmt.Errorf(err.Error()), resp
	}
	defer resp.Body.Close()

	return nil, resp
}

//
func UpdateImsiDeviceGroup(imsi *uint64) error {
	log.Info("Calling UpdateImsiDeviceGroup...")

	// Get the current configuration from the ROC
	origVal, err := migration.GetPath("", *aetherConfigTarget, *aetherConfigAddr, context.Background())
	if err != nil {
		log.Error("Failed to get the current state from onos-config: %v", err)
	}

	// Convert the JSON config into a Device structure
	origJsonBytes := origVal.GetJsonVal()
	device := &models.Device{}
	if len(origJsonBytes) > 0 {
		if err := models.Unmarshal(origJsonBytes, device); err != nil {
			log.Error("Failed to unmarshal json")
			return fmt.Errorf("Failed to unmarshal json")
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
		return AddImsiToDefaultGroup(device, dgroup, imsi)
	}
	dgroup := *site.Id + "-default"
	return AddImsiToDefaultGroup(device, dgroup, imsi)
}

//Add Imsi to default group expect the group already exists
func AddImsiToDefaultGroup(device *models.Device, dgroup string, imsi *uint64) error {
	log.Infof("AddImsiToDefaultGroup Name : %s", dgroup)

	// Now get the device group the caller wants us to add the IMSI to
	dg, okay := device.DeviceGroup.DeviceGroup[dgroup]
	if !okay {
		log.Error("Failed to find device group %v", dgroup)
		return fmt.Errorf("Failed to find device group %v", dgroup)
	}
	site, err := getDeviceGroupSite(device, dg)
	if err != nil {
		log.Error("Failed to find site for device group %v", *dg.Id)
		return fmt.Errorf("Failed to find site for device group %v", *dg.Id)
	}
	maskedImsi, err := sync.MaskSubscriberImsiDef(site.ImsiDefinition, *imsi) // mask off the MCC/MNC/EntId
	if err != nil {
		log.Error("Failed to mask the subscriber: %v", err)
		return fmt.Errorf("Failed to mask the subscriber: %v", err)
	}

	log.Infof("Masked imsi is %v", maskedImsi)

	// An imsi-range inside of a devicegroup needs a name. Let's just name our range after the imsi
	// we're creating, prepended with "auto-" to make it clear it was automatically added. Don't worry
	// about coalescing ranges -- just create simple ranges with exactly one imsi per range.
	rangeName := fmt.Sprintf("auto-%d", *imsi)

	// Generate a prefix into the gNMI configuration tree
	prefix := migration.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]/imsis[name=%s]", dgroup, rangeName), *aetherConfigTarget)

	// Build up a list of gNMI updates to apply
	updates := []*gpb.Update{}
	updates = migration.AddUpdate(updates, migration.UpdateUInt64("imsi-range-from", *aetherConfigTarget, &maskedImsi))

	// Apply them
	err = migration.Update(prefix, *aetherConfigTarget, *aetherConfigAddr, updates, context.Background())
	if err != nil {
		log.Errorf("Error executing gNMI: %v", err)
		return fmt.Errorf("Error executing gNMI: %v", err)
	}
	return nil
}

//Get site for the device group
func getDeviceGroupSite(device *models.Device, dg *models.DeviceGroup_DeviceGroup_DeviceGroup) (*models.Site_Site_Site, error) {
	if (dg.Site == nil) || (*dg.Site == "") {
		return nil, fmt.Errorf("DeviceGroup %s has no site.", *dg.Id)
	}
	site, okay := device.Site.Site[*dg.Site]
	if !okay {
		return nil, fmt.Errorf("DeviceGroup %s site %s not found.", *dg.Id, *dg.Site)
	}
	if (site.Enterprise == nil) || (*site.Enterprise == "") {
		return nil, fmt.Errorf("DeviceGroup %s has no enterprise.", *dg.Id)
	}
	return site, nil
}

//Check if any DeviceGroups contains this imsi
func findImsiInDeviceGroup(device *models.Device, imsi uint64) *models.DeviceGroup_DeviceGroup_DeviceGroup {
	log.Info("findImsiInDeviceGroup...")
deviceGroupLoop:
	for _, dg := range device.DeviceGroup.DeviceGroup {
		for _, imsiBlock := range dg.Imsis {
			site, err := getDeviceGroupSite(device, dg)
			if err != nil {
				log.Warnf("Error getting site: %v", err)
				continue deviceGroupLoop
			}

			if imsiBlock.ImsiRangeFrom == nil {
				log.Infof("imsiBlock %s in dg %s has blank ImsiRangeFrom", *imsiBlock.Name, *dg.Id)
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

//Check site for this imsi
func findSiteForTheImsi(device *models.Device, imsi uint64) *models.Site_Site_Site {

	for _, site := range device.Site.Site {

		// Check for default site
		if *site.Id == "defaultent-defaultsite" {
			continue
		}

		//Get the prefix 9 digits
		maskedImsi, err := sync.MaskSubscriberImsiDef(site.ImsiDefinition, imsi) // mask off the MCC/MNC/EntId
		if err != nil {
			log.Error("Failed to mask the subscriber: %v", err)
		}

		siteImsiValue, err := sync.FormatImsi(*site.ImsiDefinition.Format, *site.ImsiDefinition.Mcc,
			*site.ImsiDefinition.Mnc, *site.ImsiDefinition.Enterprise, maskedImsi)
		if err != nil {
			log.Error("Failed to mask the subscriber: %v", err)
		}

		log.Info("Calculated imsiValue for this site : ", siteImsiValue)

		if imsi == siteImsiValue {
			log.Info("Found the site for imsi : ", *site.Id)
			return site
		}

	}
	return nil
}

// Main
func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
	router := gin.New()
	router.Use(getlogger(), gin.Recovery())
	router.POST(SubscriberAPISuffix+":ueId", getlogger(), AddSubscriberByID)
	err := router.Run("0.0.0.0" + *bindPort)
	if err != nil {
		log.Error("Failed to start the Subscriber-Proxy %v", err.Error())
		return
	}
}

func getlogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		// Process request
		c.Next()
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()
		if raw != "" {
			path = path + "?" + raw
		}
		log.Infof("| %3d | %15s | %-7s | %s | %s",
			statusCode, clientIP, method, path, errorMessage)
	}
}

func getJsonResponse(msg string) []byte {
	var responseData response
	responseData.Status = msg
	jsonData, err := json.Marshal(responseData)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return jsonData
}
