// SPDX-License-Identifier: GPL-3.0-or-later
package api

//go:generate oapi-codegen -generate types,skip-prune  -o openapi_types.gen.go  -package api flamenco-openapi.yaml
//go:generate oapi-codegen -generate server,skip-prune -o openapi_server.gen.go -package api flamenco-openapi.yaml
//go:generate oapi-codegen -generate spec,skip-prune   -o openapi_spec.gen.go   -package api flamenco-openapi.yaml
//go:generate oapi-codegen -generate client,skip-prune -o openapi_client.gen.go -package api flamenco-openapi.yaml
