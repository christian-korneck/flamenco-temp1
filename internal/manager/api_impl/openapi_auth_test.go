// Package api_impl implements the OpenAPI API from pkg/api/flamenco-manager.yaml.
package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceURLPathVariables(t *testing.T) {
	// Test the regexp first.
	assert.True(t, urlVariablesReplacer.Match([]byte("/:var")))
	assert.True(t, urlVariablesReplacer.Match([]byte("/:var/")))

	assert.Equal(t, "", replaceURLPathVariables(""))
	assert.Equal(t, "/just/some/path", replaceURLPathVariables("/just/some/path"))
	assert.Equal(t, "/variable/at/{end}", replaceURLPathVariables("/variable/at/:end"))
	assert.Equal(t, "/mid/{var}/end", replaceURLPathVariables("/mid/:var/end"))
}
