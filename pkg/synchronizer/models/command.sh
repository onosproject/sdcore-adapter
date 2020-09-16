#!/bin/bash

# Copy the models from the config-models repository
curl https://raw.githubusercontent.com/onosproject/config-models/master/modelplugin/aether-1.0.0/aether_1_0_0/generated.go > generated.go
sed -i 's/package aether_1_0_0/package models/g' generated.go
