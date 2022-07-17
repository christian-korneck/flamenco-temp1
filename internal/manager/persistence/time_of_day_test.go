package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var emptyToD = TimeOfDay{timeOfDayNoValue, timeOfDayNoValue}

func TestIsBefore(t *testing.T) {
	test := func(expect bool, hour1, min1, hour2, min2 int) {
		time1 := TimeOfDay{hour1, min1}
		time2 := TimeOfDay{hour2, min2}

		assert.Equal(t, expect, time1.IsBefore(time2))
	}
	test(false, 0, 0, 0, 0)
	test(true, 0, 0, 0, 1)
	test(true, 1, 59, 2, 0)
	test(true, 1, 2, 1, 3)
	test(true, 1, 2, 15, 1)
	test(false, 17, 0, 8, 0)
}

func TestIsAfter(t *testing.T) {
	test := func(expect bool, hour1, min1, hour2, min2 int) {
		time1 := TimeOfDay{hour1, min1}
		time2 := TimeOfDay{hour2, min2}

		assert.Equal(t, expect, time1.IsAfter(time2))
	}
	test(false, 0, 0, 0, 0)
	test(true, 0, 1, 0, 0)
	test(true, 2, 1, 1, 59)
	test(true, 1, 3, 1, 2)
	test(true, 15, 1, 1, 2)
	test(false, 8, 0, 17, 0)
}

func TestOnDate(t *testing.T) {
	theDate := time.Date(2018, 12, 13, 7, 59, 43, 123, time.Local)
	tod := TimeOfDay{16, 47}
	expect := time.Date(2018, 12, 13, 16, 47, 0, 0, time.Local)
	assert.Equal(t, expect, tod.OnDate(theDate))

	// Midnight on the same day.
	tod = TimeOfDay{0, 0}
	expect = time.Date(2018, 12, 13, 0, 0, 0, 0, time.Local)
	assert.Equal(t, expect, tod.OnDate(theDate))

	// Midnight a day later.
	tod = TimeOfDay{24, 0}
	expect = time.Date(2018, 12, 14, 0, 0, 0, 0, time.Local)
	assert.Equal(t, expect, tod.OnDate(theDate))

}

func TestValue(t *testing.T) {
	// Test zero -> "00:00"
	tod := TimeOfDay{}
	if value, err := tod.Value(); assert.NoError(t, err) {
		assert.Equal(t, "00:00", value)
	}

	// Test 22:47 -> "22:47"
	tod = TimeOfDay{22, 47}
	if value, err := tod.Value(); assert.NoError(t, err) {
		assert.Equal(t, "22:47", value)
	}

	// Test empty -> ""
	tod = emptyToD
	if value, err := tod.Value(); assert.NoError(t, err) {
		assert.Equal(t, "", value)
	}
}

func TestScan(t *testing.T) {
	// Test zero -> empty
	tod := TimeOfDay{}
	if assert.NoError(t, tod.Scan("")) {
		assert.Equal(t, emptyToD, tod)
	}

	// Test 22:47 -> empty
	tod = TimeOfDay{22, 47}
	if assert.NoError(t, tod.Scan("")) {
		assert.Equal(t, emptyToD, tod)
	}

	// Test 22:47 -> 12:34
	tod = TimeOfDay{22, 47}
	if assert.NoError(t, tod.Scan("12:34")) {
		assert.Equal(t, TimeOfDay{12, 34}, tod)
	}

	// Test empty -> empty
	tod = emptyToD
	if assert.NoError(t, tod.Scan("")) {
		assert.Equal(t, emptyToD, tod)
	}

	// Test empty -> 12:34
	tod = emptyToD
	if assert.NoError(t, tod.Scan("12:34")) {
		assert.Equal(t, TimeOfDay{12, 34}, tod)
	}
}

func TestHasValue(t *testing.T) {
	zeroTod := TimeOfDay{}
	assert.True(t, zeroTod.HasValue(), "zero value should be midnight, and thus be a valid value")

	fullToD := TimeOfDay{22, 47}
	assert.True(t, fullToD.HasValue())

	noValueToD := TimeOfDay{timeOfDayNoValue, timeOfDayNoValue}
	assert.False(t, noValueToD.HasValue())

	onlyMinuteValue := TimeOfDay{timeOfDayNoValue, 47}
	assert.False(t, onlyMinuteValue.HasValue())

	onlyHourValue := TimeOfDay{22, timeOfDayNoValue}
	assert.False(t, onlyHourValue.HasValue())
}
