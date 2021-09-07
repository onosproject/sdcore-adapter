// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package subproxy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	models "github.com/onosproject/config-models/modelplugin/aether-3.0.0/aether_3_0_0"
	sync "github.com/onosproject/sdcore-adapter/pkg/synchronizer/v3"
	"net/http"
	"time"
)

//get logger
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

// get JSON from
func getJSONResponse(msg string) []byte {
	var responseData response
	responseData.Status = msg
	jsonData, err := json.Marshal(responseData)
	if err != nil {
		log.Error(err.Error())
		return nil
	}
	return jsonData
}

//Check site for this imsi
func findSiteForTheImsi(device *models.Device, imsi uint64) *models.Site_Site_Site {
	for _, site := range device.Site.Site {

		// Check for default site
		if *site.Id == "defaultent-defaultsite" {
			continue
		}

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

//Check if any DeviceGroups contains this imsi
func findImsiInDeviceGroup(device *models.Device, imsi uint64) *models.DeviceGroup_DeviceGroup_DeviceGroup {
	log.Info("findImsiInDeviceGroup...")
deviceGroupLoop:
	for _, dg := range device.DeviceGroup.DeviceGroup {
		for _, imsiBlock := range dg.Imsis {
			site, err := getSiteForDeviceGrp(device, dg)
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

//Get site for the device group
func getSiteForDeviceGrp(device *models.Device, dg *models.DeviceGroup_DeviceGroup_DeviceGroup) (*models.Site_Site_Site, error) {
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

// PostToWebConsole will Call webui API for subscriber provision on the SD-Core
func PostToWebConsole(postURI string, payload []byte, postTimeout time.Duration) (*http.Response, error) {
	log.Info("Calling WebUI API...")
	req, err := http.NewRequest("POST", postURI, bytes.NewBuffer(payload))
	if err != nil {
		log.Info("Error while connecting webui ", err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := Client.Do(req)
	if err != nil {
		log.Error(err.Error())
		return resp, fmt.Errorf(err.Error())
	}
	defer resp.Body.Close()

	return resp, nil
}
