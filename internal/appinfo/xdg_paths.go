package appinfo

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"os"
	"path"
	"path/filepath"

	"github.com/adrg/xdg"
)

// InFlamencoHome returns the filename in the 'flamenco home' dir, and ensures
// that the directory exists.
func InFlamencoHome(filename string) (string, error) {
	flamencoHome := os.Getenv("FLAMENCO_HOME")
	if flamencoHome == "" {
		return xdg.DataFile(path.Join(xdgApplicationName, filename))
	}
	if err := os.MkdirAll(flamencoHome, os.ModePerm); err != nil {
		return "", err
	}
	return filepath.Join(flamencoHome, filename), nil
}
