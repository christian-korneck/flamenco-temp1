// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.9.0 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"AA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
