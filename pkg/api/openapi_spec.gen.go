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
