package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
)

func mockedClock(t *testing.T) *clock.Mock {
	c := clock.NewMock()
	now, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	assert.NoError(t, err)
	c.Set(now)
	return c
}
