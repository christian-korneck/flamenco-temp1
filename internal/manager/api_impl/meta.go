// Package api_impl implements the OpenAPI API from pkg/api/flamenco-manager.yaml.
package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"net/http"

	"git.blender.org/flamenco/internal/appinfo"
	"git.blender.org/flamenco/pkg/api"
	"github.com/labstack/echo/v4"
)

func (f *Flamenco) GetVersion(e echo.Context) error {
	return e.JSON(http.StatusOK, api.FlamencoVersion{
		Version: appinfo.ApplicationVersion,
		Name:    appinfo.ApplicationName,
	})
}

func (f *Flamenco) GetConfiguration(e echo.Context) error {
	return e.JSON(http.StatusOK, api.ManagerConfiguration{
		ShamanEnabled:   f.isShamanEnabled(),
		StorageLocation: f.config.EffectiveStoragePath(),
	})
}
