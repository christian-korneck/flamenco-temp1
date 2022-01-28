package worker

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
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v2"
)

var (
	errURLWithoutHostName = errors.New("manager URL should contain a host name")
)

// WorkerConfig represents the configuration of a single worker.
// It does not include authentication credentials.
type WorkerConfig struct {
	Manager   string   `yaml:"manager_url"`
	TaskTypes []string `yaml:"task_types"`
}

type workerCredentials struct {
	WorkerID string `yaml:"worker_id"`
	Secret   string `yaml:"worker_secret"`
}

// ConfigWrangler makes it simple to load and write configuration files.
type ConfigWrangler interface {
	DefaultConfig() WorkerConfig
	WriteConfig(filename string, filetype string, config interface{}) error
	LoadConfig(filename string, config interface{}) error
}

// FileConfigWrangler is the default config wrangler that actually reads & writes files.
type FileConfigWrangler struct{}

// NewConfigWrangler returns a new ConfigWrangler instance of the default type FileConfigWrangler.
func NewConfigWrangler() ConfigWrangler {
	return FileConfigWrangler{}
}

// DefaultConfig returns a fairly sane default configuration.
func (fcw FileConfigWrangler) DefaultConfig() WorkerConfig {
	return WorkerConfig{
		Manager:   "",
		TaskTypes: []string{"sleep", "blender-render", "file-management", "exr-merge", "debug"},
	}
}

// WriteConfig stores a struct as YAML file.
func (fcw FileConfigWrangler) WriteConfig(filename string, filetype string, config interface{}) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	tempFilename := filename + "~"
	f, err := os.OpenFile(tempFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	fmt.Fprintf(f, "# %s file for Flamenco Worker.\n", filetype)
	fmt.Fprintln(f, "# For an explanation of the fields, refer to flamenco-worker-example.yaml")
	fmt.Fprintln(f, "#")
	fmt.Fprintln(f, "# NOTE: this file can be overwritten by Flamenco Worker.")
	fmt.Fprintln(f, "#")
	now := time.Now()
	fmt.Fprintf(f, "# This file was written on %s\n\n", now.Format("2006-01-02 15:04:05 -07:00"))

	n, err := f.Write(data)
	if err != nil {
		f.Close() // ignore errors here
		return err
	}
	if n < len(data) {
		f.Close() // ignore errors here
		return io.ErrShortWrite
	}
	if err = f.Close(); err != nil {
		return err
	}

	log.Debug().Str("filename", tempFilename).Msg("config file written")
	log.Debug().
		Str("from", tempFilename).
		Str("to", filename).
		Msg("renaming config file")
	if err := os.Rename(tempFilename, filename); err != nil {
		return err
	}
	log.Info().Str("filename", filename).Msg("Saved configuration file")

	return nil
}

// LoadConfig loads a YAML configuration file into 'config'
func (fcw FileConfigWrangler) LoadConfig(filename string, config interface{}) error {
	log.Debug().Str("filename", filename).Msg("loading config file")
	f, err := os.OpenFile(filename, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()

	dec := yaml.NewDecoder(f)
	if err = dec.Decode(config); err != nil {
		return err
	}

	return nil
}

// ParseURL allows URLs without scheme (assumes HTTP).
func ParseURL(rawURL string) (*url.URL, error) {
	var err error
	var parsedURL *url.URL

	parsedURL, err = url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	// url.Parse() is a bit weird when there is no scheme.
	if parsedURL.Host == "" && parsedURL.Path != "" {
		// This case happens when you just enter a hostname, like manager='thehost'
		parsedURL.Host = parsedURL.Path
		parsedURL.Path = "/"
	}
	if parsedURL.Host == "" && parsedURL.Scheme != "" && parsedURL.Opaque != "" {
		// This case happens when you just enter a hostname:port, like manager='thehost:8083'
		parsedURL.Host = parsedURL.Scheme + ":" + parsedURL.Opaque
		parsedURL.Opaque = ""
		parsedURL.Scheme = "http"
	}
	if parsedURL.Scheme == "" {
		parsedURL.Scheme = "http"
	}
	if parsedURL.Host == "" {
		return nil, errURLWithoutHostName
	}

	return parsedURL, nil
}
