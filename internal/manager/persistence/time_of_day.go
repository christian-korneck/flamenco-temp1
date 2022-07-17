package persistence

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"database/sql/driver"
	"fmt"
	"time"
)

const (
	timeOfDayStringFormat = "%02d:%02d"

	// Assigned to the Hour and Minute fields to indicate "no value".
	timeOfDayNoValue = -1
)

// TimeOfDay represents a time of day, and can be converted to/from a string.
// Its date and timezone components are ignored, and the time is supposed to be
// interpreted as local time on any date (f.e. a scheduled sleep time of some
// Worker on a certain day-of-week & local timezone).
//
// TimeOfDay structs can also represent "no value", which will be marshaled as
// an empty string.
type TimeOfDay struct {
	Hour   int
	Minute int
}

// MakeTimeOfDay converts a time.Time into a TimeOfDay.
func MakeTimeOfDay(someTime time.Time) TimeOfDay {
	return TimeOfDay{someTime.Hour(), someTime.Minute()}
}

// EmptyTimeOfDay returns a TimeOfDay struct with no value.
// See `TimeOfDay.HasValue()`.
func EmptyTimeOfDay() TimeOfDay {
	return TimeOfDay{Hour: timeOfDayNoValue, Minute: timeOfDayNoValue}
}

// Value converts a TimeOfDay to a value usable by SQL databases.
func (ot TimeOfDay) Value() (driver.Value, error) {
	var asString = ot.String()
	return asString, nil
}

// Scan updates this TimeOfDay from the value stored in a database.
func (ot *TimeOfDay) Scan(value interface{}) error {
	b, ok := value.(string)
	if !ok {
		return fmt.Errorf("expected string, received %T", value)
	}
	return ot.setString(string(b))
}

// Equals returns True iff both times represent the same time of day.
func (ot TimeOfDay) Equals(other TimeOfDay) bool {
	return ot.Hour == other.Hour && ot.Minute == other.Minute
}

// IsBefore returns True iff ot is before other.
// Ignores everything except hour and minute fields.
func (ot TimeOfDay) IsBefore(other TimeOfDay) bool {
	if ot.Hour != other.Hour {
		return ot.Hour < other.Hour
	}
	return ot.Minute < other.Minute
}

// IsAfter returns True iff ot is after other.
// Ignores everything except hour and minute fields.
func (ot TimeOfDay) IsAfter(other TimeOfDay) bool {
	if ot.Hour != other.Hour {
		return ot.Hour > other.Hour
	}
	return ot.Minute > other.Minute
}

// OnDate returns the time of day in the local timezone on the given date.
func (ot TimeOfDay) OnDate(date time.Time) time.Time {
	year, month, day := date.Date()
	return time.Date(year, month, day, ot.Hour, ot.Minute, 0, 0, time.Local)
}

func (ot TimeOfDay) String() string {
	if !ot.HasValue() {
		return ""
	}
	return fmt.Sprintf(timeOfDayStringFormat, ot.Hour, ot.Minute)
}

func (ot TimeOfDay) HasValue() bool {
	return ot.Hour != timeOfDayNoValue && ot.Minute != timeOfDayNoValue
}

func (ot *TimeOfDay) setString(value string) error {
	scanned := TimeOfDay{}
	if value == "" {
		*ot = TimeOfDay{timeOfDayNoValue, timeOfDayNoValue}
		return nil
	}

	_, err := fmt.Sscanf(value, timeOfDayStringFormat, &scanned.Hour, &scanned.Minute)
	if err != nil {
		return err
	}
	*ot = scanned
	return nil
}
