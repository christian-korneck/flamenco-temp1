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
