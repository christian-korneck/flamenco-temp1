package persistence

/* ***** BEGIN GPL LICENSE BLOCK *****
 *
 * Original Code Copyright (C) 2022 Blender Foundation.
 *
 * This file is part of Flamenco.
 *
 * Flamenco is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * Flamenco is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
 * A PARTICULAR PURPOSE.  See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * Flamenco.  If not, see <https://www.gnu.org/licenses/>.
 *
 * ***** END GPL LICENSE BLOCK ***** */

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type Job struct {
	gorm.Model
	UUID string `gorm:"type:char(36)"`

	Name     string `gorm:"type:varchar(64)"`
	JobType  string `gorm:"type:varchar(32)"`
	Priority int8   `gorm:"type:smallint"`
	Status   string `gorm:"type:varchar(32)"` // See JobStatusXxxx consts in openapi_types.gen.go

	Settings JobSettings `gorm:"type:jsonb"`
	Metadata JobMetadata `gorm:"type:jsonb"`
}

type JobSettings map[string]interface{}

func (js JobSettings) Value() (driver.Value, error) {
	return json.Marshal(js)
}
func (js *JobSettings) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &js)
}

type JobMetadata map[string]string

func (js JobMetadata) Value() (driver.Value, error) {
	return json.Marshal(js)
}
func (js *JobMetadata) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &js)
}
