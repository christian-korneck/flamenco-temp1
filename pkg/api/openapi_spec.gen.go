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
