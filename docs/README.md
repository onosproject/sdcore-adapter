**Table of Contents**

- [1. SD-Core Adapter](#1-SD-Core-Adapter)
  - [1.1. Simulator mode](#11-Simulator-mode)
- [2. Client tools for testing](#2-Client-tools-for-testing)
  - [2.1. gNMI Client User Manual](#21-gNMI-Client-User-Manual)
  - [2.2. gNOI Client User Manual](#22-gNOI-Client-User-Manual)

# 1. SD-Core Adapter

This is a docker VM that runs a gNMI and/or gNOI implementation
supporting aether SD-Core models.

Inspired by https://github.com/onosproject/gnxi-simulators and https://github.com/faucetsdn/gnmi

## 1.1. Simulator mode
The adapter can operate in three modes, controlled
using **SIM_MODE** environment variable in the docker-compose file.
1) SIM_MODE=1 as gNMI target only. The configuration is loaded by default from [configs/target_configs/empty_config.json](../configs/target_configs/empty_config.json)
2) SIM_MODE=2 as gNOI target only. It supports *Certificate management* that can be used for certificate installation and rotation.
3) SIM_MODE=3 both gNMI and gNOsI targets simultaneously

Only SIM_MODE=1 is currently implemented.

# 2. Client tools for testing
You can access to the information about client tools for each SIM_MODE
including troubleshooting tips using the following links:

## 2.1. gNMI Client User Manual
[gNMI Client_User Manual](https://github.com/onosproject/gnxi-simulators/blob/master/docs/gnmi/gnmi_user_manual.md)

## 2.2. gNOI Client User Manual
[gNOI Client_User Manual](https://github.com/onosproject/gnxi-simulators/blob/master/docs/gnoi/gnoi_user_manual.md)
