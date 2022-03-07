package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseURL(t *testing.T) {
	test := func(expected, input string) {
		actualURL, err := ParseURL(input)
		assert.Nil(t, err)
		assert.Equal(t, expected, actualURL.String())
	}

	test("http://jemoeder:1234", "jemoeder:1234")
	test("http://jemoeder/", "jemoeder")
	test("opjehoofd://jemoeder:4213/xxx", "opjehoofd://jemoeder:4213/xxx")
}
