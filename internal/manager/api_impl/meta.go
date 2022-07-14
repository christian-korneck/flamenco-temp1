// Package api_impl implements the OpenAPI API from pkg/api/flamenco-openapi.yaml.
package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"git.blender.org/flamenco/internal/appinfo"
	"git.blender.org/flamenco/internal/manager/config"
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
	isFirstRun, err := f.config.IsFirstRun()
	if err != nil {
		logger := requestLogger(e)
		logger.Error().Err(err).Msg("error investigating configuration")
		return sendAPIError(e, http.StatusInternalServerError, "error investigating configuration: %v", err)
	}

	return e.JSON(http.StatusOK, api.ManagerConfiguration{
		ShamanEnabled:   f.isShamanEnabled(),
		StorageLocation: f.config.EffectiveStoragePath(),
		IsFirstRun:      isFirstRun,
	})
}

func (f *Flamenco) GetConfigurationFile(e echo.Context) error {
	config := f.config.Get()
	return e.JSON(http.StatusOK, config)
}

func (f *Flamenco) GetVariables(e echo.Context, audience api.ManagerVariableAudience, platform string) error {
	variables := f.config.ResolveVariables(
		config.VariableAudience(audience),
		config.VariablePlatform(platform),
	)

	apiVars := api.ManagerVariables{
		AdditionalProperties: make(map[string]api.ManagerVariable),
	}
	for name, variable := range variables {
		apiVars.AdditionalProperties[name] = api.ManagerVariable{
			IsTwoway: variable.IsTwoWay,
			Value:    variable.Value,
		}
	}

	return e.JSON(http.StatusOK, apiVars)
}

func (f *Flamenco) CheckSharedStoragePath(e echo.Context) error {
	logger := requestLogger(e)

	var toCheck api.CheckSharedStoragePathJSONBody
	if err := e.Bind(&toCheck); err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}

	path := toCheck.Path
	logger = logger.With().Str("path", path).Logger()
	logger.Info().Msg("checking whether this path is suitable as shared storage")

	mkError := func(cause string, args ...interface{}) error {
		if len(args) > 0 {
			cause = fmt.Sprintf(cause, args...)
		}

		logger.Warn().Str("cause", cause).Msg("shared storage path check failed")
		return e.JSON(http.StatusOK, api.PathCheckResult{
			Cause:    cause,
			IsUsable: false,
			Path:     path,
		})
	}

	// Check for emptyness.
	if path == "" {
		return mkError("An empty path is never suitable as shared storage")
	}

	// Check whether it is actually a directory.
	stat, err := os.Stat(path)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		return mkError("This path does not exist. Choose an existing directory.")
	case err != nil:
		logger.Error().Err(err).Msg("error checking filesystem")
		return mkError("Error checking filesystem: %v", err)
	case !stat.IsDir():
		return mkError("The given path is not a directory. Choose an existing directory.")
	}

	// Check if this is the Flamenco directory itself.
	myDir, err := flamencoManagerDir()
	if err != nil {
		logger.Error().Err(err).Msg("error trying to find my own directory")
	} else if path == myDir {
		return mkError("Don't pick the installation directory of Flamenco Manager. Choose a directory dedicated to the shared storage of files.")
	}

	// See if we can create a file there.
	file, err := os.CreateTemp(path, "flamenco-writability-test-*.txt")
	if err != nil {
		return mkError("Unable to create a file in that directory: %v. "+
			"Pick an existing directory where Flamenco Manager can create files.", err)
	}

	defer func() {
		// Clean up after the test is done.
		file.Close()
		os.Remove(file.Name())
	}()

	if _, err := file.Write([]byte("Ünicöde")); err != nil {
		return mkError("unable to write to %s: %v", file.Name(), err)
	}
	if err := file.Close(); err != nil {
		// Some write errors only get reported when the file is closed, so just
		// report is as a regular write error.
		return mkError("unable to write to %s: %v", file.Name(), err)
	}

	// There is a directory, and we can create a file there. Should be good to go.
	return e.JSON(http.StatusOK, api.PathCheckResult{
		Cause:    "Directory checked OK!",
		IsUsable: true,
		Path:     path,
	})
}

func flamencoManagerDir() (string, error) {
	exename, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Dir(exename), nil
}
