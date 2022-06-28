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
	"eXvz/wMAAP//xXvKGjWoAAA=",
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
