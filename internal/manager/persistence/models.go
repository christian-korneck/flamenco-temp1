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
	UUID string `gorm:"type:char(36);not null;unique;index"`

	Name     string `gorm:"type:varchar(64);not null"`
	JobType  string `gorm:"type:varchar(32);not null"`
	Priority int8   `gorm:"type:smallint;not null"`
	Status   string `gorm:"type:varchar(32);not null"` // See JobStatusXxxx consts in openapi_types.gen.go

	Settings JobSettings     `gorm:"type:jsonb"`
	Metadata StringStringMap `gorm:"type:jsonb"`
}

type JobSettings map[string]interface{}
type StringStringMap map[string]string

type Task struct {
	gorm.Model

	Name     string `gorm:"type:varchar(64);not null"`
	Type     string `gorm:"type:varchar(32);not null"`
	JobID    uint   `gorm:"not null"`
	Job      *Job   `gorm:"foreignkey:JobID;references:ID;constraint:OnDelete:CASCADE;not null"`
	Priority int    `gorm:"type:smallint;not null"`
	Status   string `gorm:"type:varchar(16);not null"`

	// TODO: include info about which worker is/was working on this.

	// Dependencies are tasks that need to be completed before this one can run.
	Dependencies []*Task `gorm:"many2many:task_dependencies;"`

	Commands Commands `gorm:"type:jsonb"`
}

type Commands []Command

type Command struct {
	Type       string          `json:"type"`
	Parameters StringStringMap `json:"parameters"`
}

func (c Commands) Value() (driver.Value, error) {
	return json.Marshal(c)
}
func (c *Commands) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &c)
}

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

func (js StringStringMap) Value() (driver.Value, error) {
	return json.Marshal(js)
}
func (js *StringStringMap) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &js)
}
