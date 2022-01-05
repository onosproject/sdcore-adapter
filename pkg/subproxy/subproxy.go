// SPDX-FileCopyrightText: 2021-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package subproxy

import (
	"fmt"
	"github.com/gin-gonic/gin"
	models "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	"github.com/onosproject/sdcore-adapter/pkg/gnmiclient"
	sync "github.com/onosproject/sdcore-adapter/pkg/synchronizer"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc/metadata"
	"io/ioutil"
	"net/http"
	"os"
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
	log.Infof("Received One Subscriber Data")
	ueID := c.Param("ueId")
	var payload []byte
	if c.Request.Body != nil {
		payload, _ = ioutil.ReadAll(c.Request.Body)
	}

	if !strings.HasPrefix(ueID, "imsi-") {
		log.Warn("Ue Id format is invalid ")
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	log.Infof("Received subscriber id : %s ", ueID)

	split := strings.Split(ueID, "-")
	imsiValue, err := strconv.ParseUint(split[1], 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{})
		return
	}
	//Getting gnmi context
	s.gnmiContext = NewGnmiContext(c)
	err = s.updateImsiDeviceGroup(imsiValue)
	if err != nil {
		jsonByte, okay := getJSONResponse(err.Error())
		if okay != nil {
			log.Warn(err.Error())
		}
		c.Data(http.StatusInternalServerError, "application/json", jsonByte)
		return
	}

	resp, err := ForwardReqToEndpoint(s.BaseWebConsoleURL+subscriberAPISuffix+ueID, payload, s.PostTimeout)
	if err != nil {
		jsonByte, okay := getJSONResponse(err.Error())
		if okay != nil {
			log.Warn(err.Error())
		}
		c.Data(http.StatusInternalServerError, "application/json", jsonByte)
		return
	}
	if resp.StatusCode != 201 {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			jsonByte, okay := getJSONResponse(err.Error())
			if okay != nil {
				log.Warn(err.Error())
			}
			c.Data(http.StatusInternalServerError, "application/json", jsonByte)
			return
		}

		bodyBytes, err = getJSONResponse(string(bodyBytes))
		if err != nil {
			log.Warn(err.Error())
		}
		c.Data(resp.StatusCode, "application/json", bodyBytes)
		return
	}

	c.JSON(resp.StatusCode, gin.H{"status": "success"})
}

func (s *subscriberProxy) InitGnmiContext() error {

	var err error
	s.gnmiClient, s.token, err = gnmiclient.NewGnmiWithInterceptor(s.AetherConfigAddress, time.Second*15)
	if err != nil {
		log.Errorf("Error opening gNMI client %s", err.Error())
		s.gnmiClient = nil //ensure it's nil
		return err
	}
	return nil
}

func (s *subscriberProxy) getDevice() (*models.Device, error) {

	if s.gnmiClient == nil {
		err := s.InitGnmiContext()
		if err != nil {
			return nil, err
		}
	}

	//Append the auth token if oid issuer is configured
	openIDIssuer := os.Getenv("OIDC_SERVER_URL")
	if len(strings.TrimSpace(openIDIssuer)) > 0 {
		s.gnmiContext = metadata.AppendToOutgoingContext(s.gnmiContext, authorization, s.token)
	}

	//Getting Device Group only
	origValDg, err := s.gnmiClient.GetPath(s.gnmiContext, "/device-group", s.AetherConfigTarget, s.AetherConfigAddress)
	if err != nil {
		log.Error("GetPath call failed with error ", err.Error())
		//Check if the token is expired and retry with new token
		if (len(strings.TrimSpace(openIDIssuer)) > 0) && (strings.Contains(err.Error(), "expired")) {
			log.Info("Retrying with fresh token ")
			err = s.InitGnmiContext()
			if err != nil {
				return nil, err
			}
			origValDg, err = s.gnmiClient.GetPath(s.gnmiContext, "/device-group", s.AetherConfigTarget, s.AetherConfigAddress)
			if err != nil {
				return nil, errors.NewInvalid("failed to get the current state from onos-config: %v", err.Error())
			}
		} else {
			return nil, errors.NewInvalid("failed to get the current state from onos-config: %v", err.Error())
		}
	}

	//Getting Sites only
	origValSite, err := s.gnmiClient.GetPath(s.gnmiContext, "/site", s.AetherConfigTarget, s.AetherConfigAddress)
	if err != nil {
		return nil, errors.NewInvalid("failed to get the Site from onos-config: %v", err)
	}

	origValCS, err := s.gnmiClient.GetPath(s.gnmiContext, "/connectivity-service", s.AetherConfigTarget, s.AetherConfigAddress)
	if err != nil {
		return nil, errors.NewInvalid("failed to get the CS from onos-config: %v", err.Error())
	}
	log.Info("origValCS = ", string(origValCS.GetJsonVal()))

	device := &models.Device{}
	// Convert the JSON config into a Device structure for Device Group
	origJSONBytes := origValDg.GetJsonVal()
	if len(origJSONBytes) > 0 {
		if err := models.Unmarshal(origJSONBytes, device); err != nil {
			log.Error("Failed to unmarshal json", err)
			return nil, errors.NewInvalid("failed to unmarshal json", err)
		}
	}

	// Convert the JSON config into a Device structure
	origJSONBytes = origValSite.GetJsonVal()
	if len(origJSONBytes) > 0 {
		if err := models.Unmarshal(origJSONBytes, device); err != nil {
			log.Error("Failed to unmarshal json", err)
			return nil, errors.NewInvalid("failed to unmarshal json", err)
		}
	}

	// Convert the JSON config into a Device structure for Device Group
	origJSONBytes = origValCS.GetJsonVal()
	if len(origJSONBytes) > 0 {
		if err := models.Unmarshal(origJSONBytes, device); err != nil {
			log.Error("Failed to unmarshal json : ", err)
			return nil, errors.NewInvalid("failed to unmarshal json", err)
		}
	}

	baseURL := ""
	if device.ConnectivityService.ConnectivityService["cs4gtest"] != nil && device.ConnectivityService.ConnectivityService["cs4gtest"].Core_5GEndpoint != nil {
		baseURL = *device.ConnectivityService.ConnectivityService["cs4gtest"].Core_5GEndpoint
	}
	if device.ConnectivityService.ConnectivityService["cs5gtest"] != nil && device.ConnectivityService.ConnectivityService["cs5gtest"].Core_5GEndpoint != nil {
		baseURL = *device.ConnectivityService.ConnectivityService["cs5gtest"].Core_5GEndpoint
	}
	if device.ConnectivityService.ConnectivityService["aiab-cs"] != nil && device.ConnectivityService.ConnectivityService["aiab-cs"].Core_5GEndpoint != nil {
		baseURL = *device.ConnectivityService.ConnectivityService["aiab-cs"].Core_5GEndpoint
	}
	if baseURL != "" {
		s.BaseWebConsoleURL = ExtractBaseURL(baseURL)
	}
	log.Info("endpoint : ", s.BaseWebConsoleURL)
	return device, nil
}

func (s *subscriberProxy) updateImsiDeviceGroup(imsi uint64) error {

	// Getting the current configuration from the ROC for Device group and Site only.
	device, err := s.getDevice()
	if err != nil {
		return err
	}

	if device.DeviceGroup == nil {
		log.Infof("No device groups founds")
		return nil
	}

	// Check if the IMSI already exists
	dg := findImsiInDeviceGroup(device, imsi)
	if dg != nil {
		log.Infof("Imsi %v already exists in device group %s", imsi, *dg.Id)
		return nil
	}
	log.Infof("Imsi doesn't exist in any device group")

	//Check if the site exists
	site, err := findSiteForTheImsi(device, imsi)
	if err != nil {
		return err
	}
	if site == nil {
		log.Infof("No sites found for this imsi %s", imsi)
		dgroup := "defaultent-defaultsite-default"
		return s.addImsiToDefaultGroup(device, dgroup, imsi)
	}
	dgroup := *site.Id + "-default"
	return s.addImsiToDefaultGroup(device, dgroup, imsi)

}

//addImsiToDefaultGroup adds Imsi to default group expect the group already exists
func (s *subscriberProxy) addImsiToDefaultGroup(device *models.Device, dgroup string, imsi uint64) error {
	log.Infof("AddImsiToDefaultGroup Name : %s", dgroup)

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
	maskedImsi, err := sync.MaskSubscriberImsiDef(site.ImsiDefinition, imsi) // mask off the MCC/MNC/EntId
	if err != nil {
		return errors.NewInvalid("Failed to mask the subscriber: %v", err)
	}

	log.Infof("Masked imsi is %v", maskedImsi)

	// An imsi-range inside of a devicegroup needs a name. Let's just name our range after the imsi
	// we're creating, prepended with "auto-" to make it clear it was automatically added. Don't worry
	// about coalescing ranges -- just create simple ranges with exactly one imsi per range.
	rangeName := fmt.Sprintf("auto-%d", imsi)

	// Generate a prefix into the gNMI configuration tree
	prefix := gnmiclient.StringToPath(fmt.Sprintf("device-group/device-group[id=%s]/imsis[imsi-id=%s]", dgroup,
		rangeName), s.AetherConfigTarget)

	// Build up a list of gNMI updates to apply
	updates := []*gpb.Update{}
	updates = gnmiclient.AddUpdate(updates, gnmiclient.UpdateUInt64("imsi-range-from", s.AetherConfigTarget, &maskedImsi))

	// Apply them
	err = s.gnmiClient.Update(s.gnmiContext, prefix, s.AetherConfigTarget, s.AetherConfigAddress, updates)
	if err != nil {
		log.Errorf("Error while applying changes via gNMI %v", err)
		return errors.NewInternal("Error executing gNMI: %v", err)
	}
	return nil

}

//StartSubscriberProxy start the subscriber
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
	postTimeout time.Duration) *subscriberProxy {
	sproxy := &subscriberProxy{
		AetherConfigAddress: aetherConfigAddr,
		AetherConfigTarget:  aetherConfigTarget,
		BaseWebConsoleURL:   baseWebConsoleURL,
		PostTimeout:         postTimeout,
	}
	return sproxy
}
