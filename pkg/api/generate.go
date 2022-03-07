//go:generate oapi-codegen -generate types  -o openapi_types.gen.go  -package api flamenco-manager.yaml
//go:generate oapi-codegen -generate server -o openapi_server.gen.go -package api flamenco-manager.yaml
//go:generate oapi-codegen -generate spec   -o openapi_spec.gen.go   -package api flamenco-manager.yaml
//go:generate oapi-codegen -generate client -o openapi_client.gen.go -package api flamenco-manager.yaml

package api

// SPDX-License-Identifier: GPL-3.0-or-later
