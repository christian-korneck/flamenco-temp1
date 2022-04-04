package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v2"
)

var (
	errURLWithoutHostName = errors.New("manager URL should contain a host name")
)

const (
	credentialsFilename = "flamenco-worker-credentials.yaml"
	configFilename      = "flamenco-worker.yaml"
)

var defaultConfig = WorkerConfig{
	ConfiguredManager: "", // Auto-detect by default.
	TaskTypes:         []string{"blender", "ffmpeg", "file-management", "misc"},
}

// WorkerConfig represents the configuration of a single worker.
// It does not include authentication credentials.
type WorkerConfig struct {
	// ConfiguredManager is the Manager URL that's in the configuration file.
	ConfiguredManager string `yaml:"manager_url"`

	// ManagerURL is the Manager URL to use by the Worker. It could come from the
	// configuration file, but also from autodiscovery via UPnP/SSDP.
	ManagerURL string `yaml:"-"`

	TaskTypes []string `yaml:"task_types"`
}

type WorkerCredentials struct {
	WorkerID string `yaml:"worker_id"`
	Secret   string `yaml:"worker_secret"`
}

// FileConfigWrangler is the default config wrangler that actually reads & writes files.
type FileConfigWrangler struct {
	// In-memory copy of the worker configuration.
	wc    *WorkerConfig
	creds *WorkerCredentials
}

// NewConfigWrangler returns ConfigWrangler that reads files.
func NewConfigWrangler() FileConfigWrangler {
	return FileConfigWrangler{}
}

// WorkerConfig returns the worker configuration, or the default config if
// there is no config file. Configuration is only loaded from disk once;
// subsequent calls return the same config.
func (fcw *FileConfigWrangler) WorkerConfig() (WorkerConfig, error) {
	if fcw.wc != nil {
		return *fcw.wc, nil
	}

	wc := fcw.DefaultConfig()
	err := fcw.loadConfig(configFilename, &wc)

	if err != nil {
		if !os.IsNotExist(err) {
			return wc, err
		}

		// The config file not existing is fine; just use the defaults.
	}

	fcw.wc = &wc

	man := strings.TrimSpace(wc.ConfiguredManager)
	if man != "" {
		fcw.SetManagerURL(man)
	}

	return wc, nil
}

func (fcw *FileConfigWrangler) SaveConfig() error {
	err := fcw.writeConfig(configFilename, "Configuration", fcw.wc)
	if err != nil {
		return fmt.Errorf("writing to %s: %w", configFilename, err)
	}
	return nil
}

func (fcw *FileConfigWrangler) WorkerCredentials() (WorkerCredentials, error) {
	var creds WorkerCredentials
	err := fcw.loadConfig(credentialsFilename, &creds)
	if err != nil {
		return WorkerCredentials{}, err
	}

	log.Info().
		Str("filename", credentialsFilename).
		Msg("loaded credentials")
	return creds, nil
}

func (fcw *FileConfigWrangler) SaveCredentials(creds WorkerCredentials) error {
	fcw.creds = &creds

	err := fcw.writeConfig(credentialsFilename, "Credentials", creds)
	if err != nil {
		return fmt.Errorf("writing to %s: %w", credentialsFilename, err)
	}
	return nil
}

// SetManagerURL overwrites the Manager URL in the cached configuration.
// This is an in-memory change only, and will not be written to the config file.
// Returns a new copy of the WorkerConfig with the Manager URL updated.
func (fcw *FileConfigWrangler) SetManagerURL(managerURL string) WorkerConfig {
	fcw.wc.ManagerURL = managerURL
	return *fcw.wc
}

// DefaultConfig returns a fairly sane default configuration.
func (fcw FileConfigWrangler) DefaultConfig() WorkerConfig {
	return defaultConfig
}

// WriteConfig stores a struct as YAML file.
func (fcw FileConfigWrangler) writeConfig(filename string, filetype string, config interface{}) error {
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
func (fcw FileConfigWrangler) loadConfig(filename string, config interface{}) error {
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
