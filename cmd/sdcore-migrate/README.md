<!--
SPDX-FileCopyrightText: 2019-present Open Networking Foundation <info@opennetworking.org>

SPDX-License-Identifier: Apache-2.0
-->

# Migrate utility

Utility that can connect to a running Aether system, perform migration steps to produce
a migrated configuration in aether-config.

The migrated configuration output to the terminal (or to a file with the `-o` option)
in the form of a MEGA Patch that can be replayed through the ROC API.

Alternatively, using the `-out-to-gnmi` option, the output can be played back to the
source system without any translation to PATCH format.

Run with:

```bash
go run ./cmd/sdcore-migrate/sdcore-migrate.go \
--from-version 4.0.0 --from-target connectivity-service-v4 \
--to-target connectivity-service-v2 --to-version 2.0.0 \
--aether-config localhost:5150 --hostCheckDisabled \
-o /tmp/migrate_prod.json \
--ah "Bearer eyJhbGciOiJSUzI1NiIsInR5cCIgOiAiSldUIiwia2lkIiA6ICJ5eEhQbklhYV9pMjQwWTJfc00tXzRKTHFCUW5MYWNkczNYN2xEc2dRZmlVIn0.eyJleHAiOjE2MzY2NTAzMDEsImlhdCI6MTYzNjU2MzkwMiwiYXV0aF90aW1lIjoxNjM2NTYzOTAxLCJqdGkiOiI0ODZjYTNlMS1iYmVlLTQ4NzItYWFjMy1mODk1NDhjY2U0M2QiLCJpc3MiOiJodHRwOi8vazN1LWtleWNsb2FrOjU1NTcvYXV0aC9yZWFsbXMvbWFzdGVyIiwiYXVkIjoiYWV0aGVyLXJvYy1ndWkiLCJzdWIiOiI1ZGQxZjU5Yi02YTE4LTQ3ZDgtOTU5Yy1jYWM5ODhkYWJlOWYiLCJ0eXAiOiJJRCIsImF6cCI6ImFldGhlci1yb2MtZ3VpIiwibm9uY2UiOiJaV2s1TjFOSmJITXhXVUp1ZVMxS05IbElhMGxwUVhadlNqTnVMVk40TjA5a2VURi1NMWt4T1U1YWRub3ciLCJzZXNzaW9uX3N0YXRlIjoiZmI0NzlmZTYtOTdiNS00NjAyLWJmZWMtNTkxMmI4NDliZTU1IiwiYXRfaGFzaCI6InJrY1BKOVFQbXZ6c0FjTGF5S185SWciLCJhY3IiOiIxIiwic2lkIjoiZmI0NzlmZTYtOTdiNS00NjAyLWJmZWMtNTkxMmI4NDliZTU1IiwiZW1haWxfdmVyaWZpZWQiOmZhbHNlLCJuYW1lIjoiQWxpY2UgQWRtaW4iLCJncm91cHMiOlsibWl4ZWRHcm91cCIsIkFldGhlclJPQ0FkbWluIl0sInByZWZlcnJlZF91c2VybmFtZSI6ImFsaWNlYSIsImdpdmVuX25hbWUiOiJBbGljZSBBZG1pbiIsImVtYWlsIjoiYWxpY2VhQG9wZW5uZXR3b3JraW5nLm9yZyJ9.OpIPz-eLN5Aq8M-5FARQYIMWarPdTGMKSzj5VKFtGq2wFQ-VWANHK6GYYvEVx8Yn_jPfiVscH_VptB8e8COw4jlwPjfpS8ulIQC0sl5aEtEg4HUgbocI2xP4gFTqCwyt3qmCyudllBqZm6jmO5SnHNWvastU-aRggsxvsql5wsyXgOLou3couz9dI_ThjNjkgVcmJePVUVEPfvFv0eyxE3M_fmkeBju7nHGNCTxO_InV2CQQBHrPuUl2HoZWQXE4tie5SG-hBhJi4k1os-bjG1BhtreHbNppYPy7hi1BmHXlCplzO7cxS_h7i5pKEvy1RTv7dPq7YbsRUXljJwvacA"
```

> The onos-config-port:5150 can be port-forwarded to localhost with
> `kubectl -n micro-onos port-forward $(kubectl -n micro-onos get pods -l type=config -o name) 5150`

> The Bearer token is needed only if `aether-config` is operating in secure mode. The API Key can be
> got from the `aether-roc-gui`
