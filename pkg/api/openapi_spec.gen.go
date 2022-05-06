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
	"1eur/wkAAP//4wfCnT5+AAA=",
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
