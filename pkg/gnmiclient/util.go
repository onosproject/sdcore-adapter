// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

/*
 * Various utility functions for migration.
 */

package gnmiclient

import (
	"context"
	"encoding/json"
	"github.com/onosproject/onos-lib-go/pkg/errors"
	gpb "github.com/openconfig/gnmi/proto/gnmi"
	"golang.org/x/oauth2"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"net/http"
	"net/url"
	"os"
	"strings"
)

// splitKey given a string foo[k=v], return (foo, &k, &v)
// If the string does not contain a key, return (foo, nil, nil)
func splitKey(name string) (*string, map[string]string) {
	parts := strings.Split(name, "[")
	name = parts[0]
	if name == "" {
		return nil, nil
	}
	if len(parts) < 2 {
		return &name, nil
	}

	keys := make(map[string]string)
	for _, part := range parts[1:] {
		keyValue := strings.TrimRight(part, "]")
		subparts := strings.Split(keyValue, "=")
		if len(subparts) < 2 {
			return &name, nil
		}
		keys[subparts[0]] = subparts[1]
	}
	return &name, keys
}

// StringToPath converts a string for the format x/y/z[k=v] into a gdb.Path
func StringToPath(s string, target string) *gpb.Path {
	elems := []*gpb.PathElem{}

	parts := strings.Split(s, "/")

	for _, name := range parts {
		if len(name) > 0 {
			var keys map[string]string

			name, keys := splitKey(name)
			if name == nil {
				// the term was empty
				continue
			}

			elem := &gpb.PathElem{Name: *name,
				Key: keys}

			elems = append(elems, elem)
		}
	}

	return &gpb.Path{
		Target: target,
		Elem:   elems,
	}
}

// UpdateString creates a gpb.Update for a string value
func UpdateString(path string, target string, val *string) *gpb.Update {
	if val == nil {
		return nil
	}

	return &gpb.Update{
		Path: StringToPath(path, target),
		Val:  &gpb.TypedValue{Value: &gpb.TypedValue_StringVal{StringVal: *val}},
	}
}

// UpdateUInt8 creates a gpb.Update for a uint8 value
func UpdateUInt8(path string, target string, val *uint8) *gpb.Update {
	if val == nil {
		return nil
	}

	return &gpb.Update{
		Path: StringToPath(path, target),
		Val:  &gpb.TypedValue{Value: &gpb.TypedValue_UintVal{UintVal: uint64(*val)}},
	}
}

// UpdateInt8 creates a gpb.Update for a int8 value
func UpdateInt8(path string, target string, val *int8) *gpb.Update {
	if val == nil {
		return nil
	}

	return &gpb.Update{
		Path: StringToPath(path, target),
		Val:  &gpb.TypedValue{Value: &gpb.TypedValue_UintVal{UintVal: uint64(*val)}},
	}
}

// UpdateUInt16 creates a gpb.Update for a uint16 value
func UpdateUInt16(path string, target string, val *uint16) *gpb.Update {
	if val == nil {
		return nil
	}

	return &gpb.Update{
		Path: StringToPath(path, target),
		Val:  &gpb.TypedValue{Value: &gpb.TypedValue_UintVal{UintVal: uint64(*val)}},
	}
}

// UpdateUInt32 creates a gpb.Update for a uint32 value
func UpdateUInt32(path string, target string, val *uint32) *gpb.Update {
	if val == nil {
		return nil
	}

	return &gpb.Update{
		Path: StringToPath(path, target),
		Val:  &gpb.TypedValue{Value: &gpb.TypedValue_UintVal{UintVal: uint64(*val)}},
	}
}

// UpdateUInt64 creates a gpb.Update for a uint64 value
func UpdateUInt64(path string, target string, val *uint64) *gpb.Update {
	if val == nil {
		return nil
	}

	return &gpb.Update{
		Path: StringToPath(path, target),
		Val:  &gpb.TypedValue{Value: &gpb.TypedValue_UintVal{UintVal: *val}},
	}
}

// UpdateBool creates a gpb.Update for a bool value
func UpdateBool(path string, target string, val *bool) *gpb.Update {
	if val == nil {
		return nil
	}

	return &gpb.Update{
		Path: StringToPath(path, target),
		Val:  &gpb.TypedValue{Value: &gpb.TypedValue_BoolVal{BoolVal: *val}},
	}
}

// AddUpdate adds a gpb.Update to a list of updates, only if the gpb.Update is not
// nil.
func AddUpdate(updates []*gpb.Update, update *gpb.Update) []*gpb.Update {
	if update != nil {
		updates = append(updates, update)
	}
	return updates
}

// DeleteFromUpdates given a list of Updates, create a corresponding list of deletes
func DeleteFromUpdates(updates []*gpb.Update, target string) []*gpb.Path {
	deletePaths := []*gpb.Path{}
	for _, update := range updates {
		deletePaths = append(deletePaths, &gpb.Path{
			Target: target,
			Elem:   update.Path.Elem,
		})
	}
	return deletePaths
}

// OverrideFromContext if a context includes a particular key, then use it, otherwise use the default
func OverrideFromContext(ctx context.Context, key interface{}, defaultValue interface{}) interface{} {
	if contextVal := ctx.Value(key); contextVal != nil {
		// Use the key from the context, if available
		return contextVal
	}
	// Revert to the default, otherwise
	return defaultValue
}

// StrDeref safely dereference a *string for printing
func StrDeref(s *string) string {
	if s == nil {
		return "nil"
	}
	return *s
}

// fetchATokenViaKeyCloak Get the token via keycloak using curl
func fetchATokenViaKeyCloak(openIDIssuer string, user string, passwd string) (string, error) {

	data := url.Values{}
	data.Set("username", user)
	data.Set("password", passwd)
	data.Set("grant_type", "password")
	data.Set("client_id", "aether-roc-gui")
	data.Set("scope", "openid profile email offline_access groups")

	req, err := http.NewRequest("POST", openIDIssuer+"/protocol/openid-connect/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	log.Debug("Response Code : ", resp.StatusCode)

	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		target := new(oauth2.Token)
		err = json.NewDecoder(resp.Body).Decode(target)
		if err != nil {
			return "", err
		}
		log.Infof("Token info : Type : %v , Access Token : %v , Expiry %v : ", target.TokenType, target.AccessToken, target.Expiry)
		return target.AccessToken, nil
	}

	return "", errors.NewInvalid("Error HTTP response code : ", resp.StatusCode)

}

// getNamespace Get the current namespace
func getNamespace() string {
	if ns, ok := os.LookupEnv("POD_NAMESPACE"); ok {
		return ns
	}

	if data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	//default namespace aether-roc
	return "aether-roc"
}

// GetAccessToken authenticate and get the access token
func GetAccessToken(openIDIssuer string, secretName string) (string, error) {

	// creates the in-cluster config
	config, err := rest.InClusterConfig()

	// for out-cluster config  for dev testing purpose only
	//comment  the above line config, err := rest.InClusterConfig() and comment out the following section
	/*var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	*/
	//end of out-cluster-config

	if err != nil {
		panic(err)
	}
	// creates the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	//get secret client
	secretsClient := clientset.CoreV1().Secrets(getNamespace())

	ctx := context.Background()
	secret, err := secretsClient.Get(ctx, secretName, metaV1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	user := secret.Data["username"]
	passwd := secret.Data["password"]

	log.Info("Username : ", string(user))

	return fetchATokenViaKeyCloak(openIDIssuer, string(user), string(passwd))

}
