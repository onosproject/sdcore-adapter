// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package version

import (
	"github.com/onosproject/onos-lib-go/pkg/logging"
)

var log = logging.GetLogger("version")

// Default build-time variable.
// These values can (should) be overridden via ldflags when built with
// `make`
var (
	Version   = "unknown-version"
	GoVersion = "unknown-goversion"
	GitCommit = "unknown-gitcommit"
	GitDirty  = "unknown-gitdirty"
	BuildTime = "unknown-buildtime"
	Os        = "unknown-os"
	Arch      = "unknown-arch"
)

// LogVersion logs the version info
func LogVersion(indent string) {
	log.Infof("%sVersion:      %s\n", indent, Version)
	log.Infof("%sGoVersion:    %s\n", indent, GoVersion)
	log.Infof("%sGit Commit:   %s\n", indent, GitCommit)
	log.Infof("%sGit Dirty:    %s\n", indent, GitDirty)
	log.Infof("%sBuilt:        %s\n", indent, BuildTime)
	log.Infof("%sOS/Arch:      %s/%s\n", indent, Os, Arch)
}
