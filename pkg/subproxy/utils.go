// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package subproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	models "github.com/onosproject/config-models/modelplugin/aether-4.0.0/aether_4_0_0"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	sync "github.com/onosproject/sdcore-adapter/pkg/synchronizer"
	"google.golang.org/grpc/metadata"
	"net/http"
	"time"
)

const (
	authorization = "Authorization"
	host          = "Host"
	userAgent     = "User-Agent"
	remoteAddr    = "remoteAddr"
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
		log.Debugf("| %3d | %15s | %-7s | %s | %s",
			statusCode, clientIP, method, path, errorMessage)
	}
}

// get JSON from
func getJSONResponse(msg string) ([]byte, error) {
	responseData := make(map[string]interface{})
	responseData["status"] = msg
	jsonData, err := json.Marshal(responseData)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

//Check site for this imsi
func findSiteForTheImsi(device *models.Device, imsi uint64) (*models.OnfSite_Site_Site, error) {
	for _, site := range device.Site.Site {

		// Check for default site
		if *site.Id == "defaultent-defaultsite" {
			continue
		}

		maskedImsi, err := sync.MaskSubscriberImsiDef(site.ImsiDefinition, imsi) // mask off the MCC/MNC/EntId
		if err != nil {
			return nil, errors.NewInvalid("Failed to mask the subscriber: %v", err)
		}

		siteImsiValue, err := sync.FormatImsi(*site.ImsiDefinition.Format, *site.ImsiDefinition.Mcc,
			*site.ImsiDefinition.Mnc, *site.ImsiDefinition.Enterprise, maskedImsi)
		if err != nil {
			return nil, errors.NewInvalid("Failed to mask the subscriber: %v", err)
		}

		if imsi == siteImsiValue {
			log.Debugf("Found the site for imsi : ", *site.Id)
			return site, nil
		}

	}
	return nil, nil
}

//Check if any DeviceGroups contains this imsi
func findImsiInDeviceGroup(device *models.Device, imsi uint64) *models.OnfDeviceGroup_DeviceGroup_DeviceGroup {
deviceGroupLoop:
	for _, dg := range device.DeviceGroup.DeviceGroup {
		for _, imsiBlock := range dg.Imsis {
			site, err := getSiteForDeviceGrp(device, dg)
			if err != nil {
				log.Debugf("Error getting site: %v", err)
				continue deviceGroupLoop
			}

			if imsiBlock.ImsiRangeFrom == nil {
				log.Debugf("imsiBlock %s in dg %s has blank ImsiRangeFrom", *imsiBlock.ImsiId, *dg.Id)
				continue deviceGroupLoop
			}
			var firstImsi uint64
			firstImsi, err = sync.FormatImsiDef(site.ImsiDefinition, *imsiBlock.ImsiRangeFrom)
			if err != nil {
				log.Debugf("Failed to format IMSI in dg %s: %v", *dg.Id, err)
				continue deviceGroupLoop
			}
			var lastImsi uint64
			if imsiBlock.ImsiRangeTo == nil {
				lastImsi = firstImsi
			} else {
				lastImsi, err = sync.FormatImsiDef(site.ImsiDefinition, *imsiBlock.ImsiRangeTo)
				if err != nil {
					log.Debugf("Failed to format IMSI in dg %s: %v", *dg.Id, err)
					continue deviceGroupLoop
				}

			}
			log.Debugf("Compare %v %v %v", imsi, firstImsi, lastImsi)
			if (imsi >= firstImsi) && (imsi <= lastImsi) {
				return dg
			}
		}
	}
	return nil
}

//Get site for the device group
func getSiteForDeviceGrp(device *models.Device, dg *models.OnfDeviceGroup_DeviceGroup_DeviceGroup) (*models.OnfSite_Site_Site, error) {
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

// PostToWebConsole will Call webui API for subscriber provision on the SD-Core
func postToWebConsole(postURI string, payload []byte, postTimeout time.Duration) (*http.Response, error) {
	req, err := http.NewRequest("POST", postURI, bytes.NewBuffer(payload))
	if err != nil {
		return nil, errors.NewInvalid("Error while connecting webui ", err.Error())
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := clientHTTP.Do(req)
	if err != nil {
		return resp, errors.NewInvalid(err.Error())
	}
	defer resp.Body.Close()

	return resp, nil
}

// NewGnmiContext - convert the gin context in to a gRPC Context
func NewGnmiContext(httpContext *gin.Context) context.Context {

	return metadata.AppendToOutgoingContext(context.Background(),
		authorization, httpContext.Request.Header.Get(authorization),
		host, httpContext.Request.Host,
		"ua", httpContext.Request.Header.Get(userAgent), // `User-Agent` would be over written by gRPC
		remoteAddr, httpContext.Request.RemoteAddr)
}
