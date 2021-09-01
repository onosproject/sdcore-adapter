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
--from-version 2.1.0 --from-target connectivity-service-v2 \
--to-target connectivity-service-v3 --to-version 3.0.0 \
--aether-config localhost:5150 --hostCheckDisabled \
-o /tmp/migrate_prod.json \
--ah "Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6Ijc4OTZlOGYwMzYzMTQxNzgzYThkZTY5ZTQ4ZjJkODIzY2NiNmUxYmYifQ.eyJpc3MiOiJodHRwOi8vZGV4LWxkYXAtdW1icmVsbGE6NTU1NiIsInN1YiI6IkNnWmhiR2xqWldFU0JHeGtZWEEiLCJhdWQiOiJhZXRoZXItcm9jLWd1aSIsImV4cCI6MTYzMDQ0MDg3MCwiaWF0IjoxNjMwMzU0NDcwLCJub25jZSI6Ik1uaG5ZbVJaWmt4VlozaGZhVFZpVW5wbU5VcGhVMzQ0TVZKUlJYWTNOa040VjJzMllXcHFOSFpHTVhOYSIsImF0X2hhc2giOiJYVHB5NllKVEJBMEJOYkMtSGhGUHdRIiwiY19oYXNoIjoiRmRZQk5PNndUbzdpQXRzZ2FESHNHZyIsImVtYWlsIjoiYWxpY2VhQG9wZW5uZXR3b3JraW5nLm9yZyIsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJncm91cHMiOlsibWl4ZWRHcm91cCIsIkFldGhlclJPQ0FkbWluIiwiRW50ZXJwcmlzZUFkbWluIl0sIm5hbWUiOiJBbGljZSBBZG1pbiJ9.uaj481DH4BH3jYWlYwCuJfOQpci7E0pr1TPqb8t964PwNr_sv24ttM-Cxz6Nb4e35nNJFujnD5YmFd_vNVFqeaAIM0lFwX_91etptu1LaPY3BeUnTTRePyQhQktU8hIaqMd5ouWfwL62IxDc8XG02FghI2Yi2NbJc1gp6o6E6vnoYkW47n1AweqFoxUPPYFUOGfh_rH6ZlT6v3r-bBsAdKiaE1De5Q76f85mAqEWpeb2J5fFeQ_tFJKZE--yE6ad3CeSX19AfrdYAW5T3BDbS09KgVKr8q_rq8Ajz30NEFwFOyjHZ63CZuMd8Hn-F9m6pq1KeMta0h_VdEh9faEDaw"
```

> The onos-config-port:5150 can be port-forwarded to localhost with
> `kubectl -n micro-onos port-forward $(kubectl -n micro-onos get pods -l type=config -o name) 5150`

> The Bearer token is needed only if `aether-config` is operating in secure mode. The API Key can be
> got from the `aether-roc-gui`