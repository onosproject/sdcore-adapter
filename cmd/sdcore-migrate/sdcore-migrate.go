// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package main

/*
 * Main file for data migration
 *
 * Example invocation:
 *
 * /usr/local/bin/sdcore-migrate --from-target connectivity-service-v1 --to-target connectivity-service-v2 --from-version 1.0.0 --to-version 2.0.0 --aether-config onos-config:5150 -client_key=/etc/sdcore-adapter/certs/tls.key -client_crt=/etc/sdcore-adapter/certs/tls.crt -ca_crt=/etc/sdcore-adapter/certs/tls.cacert -hostCheckDisabled
 */

import (
	"flag"

	models_v1 "github.com/onosproject/config-models/modelplugin/aether-1.0.0/aether_1_0_0"
	modelplugin_v1 "github.com/onosproject/config-models/modelplugin/aether-1.0.0/modelplugin"
	models_v2 "github.com/onosproject/config-models/modelplugin/aether-2.0.0/aether_2_0_0"
	modelplugin_v2 "github.com/onosproject/config-models/modelplugin/aether-2.0.0/modelplugin"
	"github.com/onosproject/onos-lib-go/pkg/logging"
	"github.com/onosproject/sdcore-adapter/pkg/gnmi"
	"github.com/onosproject/sdcore-adapter/pkg/migration"
	"github.com/onosproject/sdcore-adapter/pkg/migration/steps"
	"github.com/openconfig/ygot/ygot"
	"reflect"
)

var (
	fromTarget       = flag.String("from-target", "", "target device to migrate from")
	toTarget         = flag.String("to-target", "", "target device to migrate to")
	fromVersion      = flag.String("from-version", "", "modeling version to migrate from")
	toVersion        = flag.String("to-version", "", "modeling version to migrate to")
	aetherConfigAddr = flag.String("aether-config", "", "address of aether-config")
)

var log = logging.GetLogger("sdcore-migrate")

func main() {
	flag.Parse()

	v1Models := gnmi.NewModel(modelplugin_v1.ModelData,
		reflect.TypeOf((*models_v1.Device)(nil)),
		models_v1.SchemaTree["Device"],
		models_v1.Unmarshal,
		//models.ΛEnum  // NOTE: There is no Enum in the aether models? So use a blank map.
		map[string]map[int64]ygot.EnumDefinition{},
	)

	v2Models := gnmi.NewModel(modelplugin_v2.ModelData,
		reflect.TypeOf((*models_v2.Device)(nil)),
		models_v2.SchemaTree["Device"],
		models_v2.Unmarshal,
		//models.ΛEnum  // NOTE: There is no Enum in the aether models? So use a blank map.
		map[string]map[int64]ygot.EnumDefinition{},
	)

	// Initialize the migration engine and register migration steps.
	mig := migration.NewMigrator(*aetherConfigAddr)
	mig.AddMigrationStep("1.0.0", v1Models, "2.0.0", v2Models, steps.MigrateV1V2)

	sec := migration.GetDefaultSecuritySettings()

	// Perform the migration
	err := mig.Migrate(*fromTarget, *fromVersion, *toTarget, *toVersion, sec)
	if err != nil {
		log.Errorf("%v", err)
	}
}
