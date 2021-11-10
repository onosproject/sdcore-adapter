# SD-Core Adapter

[![Build Status](https://api.travis-ci.org/onosproject/sdcore-adapter.svg?branch=master)](https://travis-ci.org/onosproject/gnxi-simulators)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/gojp/goreportcard/blob/master/LICENSE)
[![GoDoc](https://godoc.org/github.com/onosproject/simulators?status.svg)](https://godoc.org/github.com/onosproject/sdcore-adapter)

Implements an adapter for using SD-Core components with Aether-Config. Does the following:

* Listens for gNMI requests from Aether-Config
* Maintains an in-memory configuration store
* Creates JSON output from the configuration changes, emitting that output to log and optionally writing it to a file.

What this adapter does not do:

* Does not persistently store configuration. If the adapter is restarted, configuration will be lost. It's assumed configuration pushes can/will be retriggered through aether-config.
* Does not intelligently process diffs. Every update that occurs will cause the entire configuration JSON to be emitted. Pushing data to the southbound service is assumed to be idempotent and can be repeated multiple times with no ill effect.

It is assumed that the configuration schema at the adapter's northbound API may differ from the configuration schema of the adapter's southbound API. One of the purposes of the adapter is to translate between those two different APIs, which may evolve at different paces and may not be identical. Adapters are not general-purpose translators; They are translators written with a specific service and a specific schema in mind.

# Writing a synchronizer

"Synchronization" code is located in the `pkg/synchronizer` directory of the adapter. This code is intended to be customized for each adapter use case. The synchronizer interface exports two important methods:

* `Synchronize(config)`. Called by the GNMI server when state should be synchronized to the southbound service.
* `GetModels()`. Called by the GNMI server during startup, to determine the model schema that will be served.

When writing a new adapter, replace the `pkg/synchronizer` directory with your own. Also, rename the `cmd/sdcore-adapter` command.

# Data model migration

Migration is a series of steps that migrate from one version of the model to another. For example,
for migration from V1 to V4, the following steps would be executed in order:
* step_V3_0_0_V4_0_0

Each step has a model version and modelplugin associated with it. Migrations are performed by reading the source models, translating them to destination models, writing the destination models, and then deleting the original source models.

Migrations are executed between specific versions and targets. For example,

```bash
sdcore-migrate -from-target connectivity-v2 -from-version 2.1.0 -to-target connectivity-v3 -to-version 3.0.0  \
--aether-config localhost:5150 --hostCheckDisabled --ah "Bearer ????"
```

See [here](./cmd/sdcore-migrate/README.md) for more details.

> For now, from_target and to_target must be different in order to compensate for an issue in aether-config, but eventually the expectation is that from_target and to_target can be the same.

# Additional Documentation

[How to run](docs/README.md) SD-Core Adapter and related commands.
