//go:generate oapi-codegen -generate types  -o openapi_types.gen.go  -package api flamenco-manager.yaml
//go:generate oapi-codegen -generate server -o openapi_server.gen.go -package api flamenco-manager.yaml
//go:generate oapi-codegen -generate spec   -o openapi_spec.gen.go   -package api flamenco-manager.yaml
//go:generate oapi-codegen -generate client -o openapi_client.gen.go -package api flamenco-manager.yaml

package api

/* ***** BEGIN GPL LICENSE BLOCK *****
 *
 * Original Code Copyright (C) 2022 Blender Foundation.
 *
 * This file is part of Flamenco.
 *
 * Flamenco is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * Flamenco is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
 * A PARTICULAR PURPOSE.  See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * Flamenco.  If not, see <https://www.gnu.org/licenses/>.
 *
 * ***** END GPL LICENSE BLOCK ***** */
