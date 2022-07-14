package config

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	yaml "gopkg.in/yaml.v2"

	"git.blender.org/flamenco/internal/appinfo"
	"git.blender.org/flamenco/pkg/crosspath"
	shaman_config "git.blender.org/flamenco/pkg/shaman/config"
)

const (
	configFilename = "flamenco-manager.yaml"

	latestConfigVersion = 3

	// // relative to the Flamenco Server Base URL:
	// jwtPublicKeysRelativeURL = "api/flamenco/jwt/public-keys"
)

var (
	// ErrMissingVariablePlatform is returned when a variable doesn't declare any valid platform for a certain value.
	ErrMissingVariablePlatform = errors.New("variable's value is missing platform declaration")
	// ErrBadDirection is returned when a direction doesn't match "oneway" or "twoway"
	ErrBadDirection = errors.New("variable's direction is invalid")

	// Valid values for the "audience" tag of a ConfV2 variable.
	validAudiences = map[VariableAudience]bool{
		VariableAudienceAll:     true,
		VariableAudienceWorkers: true,
		VariableAudienceUsers:   true,
	}
)

// BlenderRenderConfig represents the configuration required for a test render.
type BlenderRenderConfig struct {
	JobStorage   string `yaml:"job_storage"`
	RenderOutput string `yaml:"render_output"`
}

// TestTasks represents the 'test_tasks' key in the Manager's configuration file.
type TestTasks struct {
	BlenderRender BlenderRenderConfig `yaml:"test_blender_render"`
}

// ConfMeta contains configuration file metadata.
type ConfMeta struct {
	// Version of the config file structure.
	Version int `yaml:"version"`
}

// Base contains those settings that are shared by all configuration versions.
type Base struct {
	Meta ConfMeta `yaml:"_meta"`

	ManagerName string `yaml:"manager_name"`
	DatabaseDSN string `yaml:"database"`
	Listen      string `yaml:"listen"`

	SSDPDiscovery bool `yaml:"autodiscoverable"`

	// LocalManagerStoragePath is where the Manager stores its files, like task
	// logs, last-rendered images, etc.
	LocalManagerStoragePath string `yaml:"local_manager_storage_path"`

	// SharedStoragePath is where files shared between Manager and Workers go,
	// like the blend files of a render job.
	SharedStoragePath string `yaml:"shared_storage_path"`

	Shaman shaman_config.Config `yaml:"shaman"`

	TaskTimeout   time.Duration `yaml:"task_timeout"`
	WorkerTimeout time.Duration `yaml:"worker_timeout"`

	/* This many failures (on a given job+task type combination) will ban a worker
	 * from that task type on that job. */
	BlocklistThreshold int `yaml:"blocklist_threshold"`

	// When this many workers have tried the task and failed, it will be hard-failed
	// (even when there are workers left that could technically retry the task).
	TaskFailAfterSoftFailCount int `yaml:"task_fail_after_softfail_count"`
}

// GarbageCollect contains the config options for the GC.
type ShamanGarbageCollect struct {
	// How frequently garbage collection is performed on the file store:
	Period time.Duration `yaml:"period"`
	// How old files must be before they are GC'd:
	MaxAge time.Duration `yaml:"maxAge"`
	// Paths to check for symlinks before GC'ing files.
	ExtraCheckoutDirs []string `yaml:"extraCheckoutPaths"`

	// Used by the -gc CLI arg to silently disable the garbage collector
	// while we're performing a manual sweep.
	SilentlyDisable bool `yaml:"-"`
}

// Conf is the latest version of the configuration.
// Currently it is version 3.
type Conf struct {
	Base `yaml:",inline"`

	// Variable name → Variable definition
	Variables map[string]Variable `yaml:"variables"`

	// Implicit variables work as regular variables, but do not get written to the
	// configuration file.
	implicitVariables map[string]Variable `yaml:"-"`

	// audience + platform + variable name → variable value.
	// Used to look up variables for a given platform and audience.
	// The 'audience' is never "all" or ""; only concrete audiences are stored here.
	VariablesLookup map[VariableAudience]map[VariablePlatform]map[string]string `yaml:"-"`
}

// Variable defines a configuration variable.
type Variable struct {
	IsTwoWay bool `yaml:"is_twoway,omitempty" json:"is_twoway,omitempty"`
	// Mapping from variable value to audience/platform definition.
	Values VariableValues `yaml:"values" json:"values"`
}

// VariableValues is the list of values of a variable.
type VariableValues []VariableValue

// VariableValue defines which audience and platform see which value.
type VariableValue struct {
	// Audience defines who will use this variable, either "all", "workers", or "users". Empty string is "all".
	Audience VariableAudience `yaml:"audience,omitempty" json:"audience,omitempty"`

	// Platforms that use this value. Only one of "Platform" and "Platforms" may be set.
	Platform  VariablePlatform   `yaml:"platform,omitempty" json:"platform,omitempty"`
	Platforms []VariablePlatform `yaml:"platforms,omitempty,flow" json:"platforms,omitempty"`

	// The actual value of the variable for this audience+platform.
	Value string `yaml:"value" json:"value"`
}

// ResolvedVariable represents info about the variable, resolved for a specific platform & audience.
type ResolvedVariable struct {
	IsTwoWay bool
	Value    string
}

// getConf parses flamenco-manager.yaml and returns its contents as a Conf object.
func getConf() (Conf, error) {
	return loadConf(configFilename)
}

// DefaultConfig returns a copy of the default configuration.
func DefaultConfig(override ...func(c *Conf)) Conf {
	c := defaultConfig
	c.Meta.Version = latestConfigVersion
	c.processAfterLoading(override...)
	return c
}

// loadConf parses the given file and returns its contents as a Conf object.
func loadConf(filename string, overrides ...func(c *Conf)) (Conf, error) {
	log.Info().Str("file", filename).Msg("loading configuration")
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		var evt *zerolog.Event
		if errors.Is(err, fs.ErrNotExist) {
			evt = log.Debug()
		} else {
			evt = log.Warn().Err(err)
		}
		evt.Msg("unable to load configuration, using defaults")
		return DefaultConfig(overrides...), err
	}

	// First parse attempt, find the version.
	baseConf := Base{}
	if err := yaml.Unmarshal(yamlFile, &baseConf); err != nil {
		return Conf{}, fmt.Errorf("unable to parse %s: %w", filename, err)
	}

	// Versioning was supported from Flamenco config v1 to v2, but not further.
	if baseConf.Meta.Version != latestConfigVersion {
		return Conf{}, fmt.Errorf(
			"configuration file %s version %d, but only version %d is supported",
			filename, baseConf.Meta.Version, latestConfigVersion)
	}

	// Second parse attempt, based on the version found.
	c := DefaultConfig()
	if err := yaml.Unmarshal(yamlFile, &c); err != nil {
		return c, fmt.Errorf("unable to parse %s: %w", filename, err)
	}

	c.processAfterLoading(overrides...)

	return c, nil
}

// processAfterLoading processes and checks the loaded config.
// This is called not just after loading from disk, but also after getting the
// default configuration.
func (c *Conf) processAfterLoading(override ...func(c *Conf)) {
	for _, overrideFunc := range override {
		overrideFunc(c)
	}

	c.processStorage()
	c.addImplicitVariables()
	c.ensureVariablesUnique()
	c.constructVariableLookupTable()
	c.checkDatabase()
	c.checkVariables()
}

func (c *Conf) processStorage() {
	// The shared storage path should be absolute, but only if it's actually configured.
	if c.SharedStoragePath != "" {
		storagePath, err := filepath.Abs(c.SharedStoragePath)
		if err != nil {
			log.Error().Err(err).
				Str("storage_path", c.SharedStoragePath).
				Msg("unable to determine absolute storage path")
		} else {
			c.SharedStoragePath = storagePath
		}
	}

	// Shaman should use the Flamenco storage location.
	c.Shaman.StoragePath = c.SharedStoragePath
}

// EffectiveStoragePath returns the absolute path of the job storage directory.
// This is made from a combination of the configured job storage path and a
// Shaman-specific subpath (if enabled).
func (c *Conf) EffectiveStoragePath() string {
	var jobStorage string
	if c.Shaman.Enabled {
		jobStorage = c.Shaman.CheckoutPath()
	} else {
		jobStorage = c.SharedStoragePath
	}

	absPath, err := filepath.Abs(jobStorage)
	if err != nil {
		log.Warn().
			Str("storagePath", jobStorage).
			Bool("shamanEnabled", c.Shaman.Enabled).
			Err(err).Msg("unable to find absolute path of storage path")
		absPath = jobStorage
	}

	return absPath
}

func (c *Conf) addImplicitVariables() {
	c.implicitVariables = make(map[string]Variable)

	c.implicitVariables["jobs"] = Variable{
		IsTwoWay: false,
		Values: []VariableValue{
			{
				Audience: VariableAudienceAll,
				Platform: VariablePlatformAll,
				Value:    crosspath.ToSlash(c.EffectiveStoragePath()),
			},
		},
	}
}

// ensureVariablesUnique erases configured variables when there are implicit
// variables with the same name.
func (c *Conf) ensureVariablesUnique() {
	for varname := range c.implicitVariables {
		if _, found := c.Variables[varname]; !found {
			continue
		}
		log.Warn().Str("variable", varname).
			Msg("configured variable will be removed, as there is an implicit variable with the same name")
		delete(c.Variables, varname)
	}
}

func (c *Conf) constructVariableLookupTable() {
	if c.VariablesLookup == nil {
		c.VariablesLookup = map[VariableAudience]map[VariablePlatform]map[string]string{}
	}

	c.constructVariableLookupTableForVars(c.Variables)
	c.constructVariableLookupTableForVars(c.implicitVariables)

	// log.Trace().
	// 	Interface("variables", c.Variables).
	// 	Msg("constructed lookup table")
}

func (c *Conf) constructVariableLookupTableForVars(vars map[string]Variable) {
	// Construct a list of all audiences except "" and "all"
	concreteAudiences := []VariableAudience{}
	isWildcard := map[VariableAudience]bool{"": true, VariableAudienceAll: true}
	for audience := range validAudiences {
		if isWildcard[audience] {
			continue
		}
		concreteAudiences = append(concreteAudiences, audience)
	}
	// log.Trace().
	// 	Interface("concreteAudiences", concreteAudiences).
	// 	Interface("isWildcard", isWildcard).
	// 	Msg("constructing variable lookup table")

	// Just for brevity.
	lookup := c.VariablesLookup

	// setValue expands wildcard audiences into concrete ones.
	var setValue func(audience VariableAudience, platform VariablePlatform, name, value string)
	setValue = func(audience VariableAudience, platform VariablePlatform, name, value string) {
		if isWildcard[audience] {
			for _, aud := range concreteAudiences {
				setValue(aud, platform, name, value)
			}
			return
		}

		if lookup[audience] == nil {
			lookup[audience] = map[VariablePlatform]map[string]string{}
		}
		if lookup[audience][platform] == nil {
			lookup[audience][platform] = map[string]string{}
		}
		// log.Trace().
		// 	Str("audience", string(audience)).
		// 	Str("platform", string(platform)).
		// 	Str("name", name).
		// 	Str("value", value).
		// 	Msg("setting variable")
		lookup[audience][platform][name] = value
	}

	// Construct the lookup table for each audience+platform+name
	for name, variable := range vars {
		// log.Trace().
		// 	Str("name", name).
		// 	Interface("variable", variable).
		// 	Msg("handling variable")
		for _, value := range variable.Values {

			// Two-way values should not end in path separator.
			// Given a variable 'apps' with value '/path/to/apps',
			// '/path/to/apps/blender' should be remapped to '{apps}/blender'.
			if variable.IsTwoWay {
				if strings.Contains(value.Value, "\\") {
					log.Warn().
						Str("variable", name).
						Str("audience", string(value.Audience)).
						Str("platform", string(value.Platform)).
						Str("value", value.Value).
						Msg("Backslash found in variable value. Change paths to use forward slashes instead.")
				}
				value.Value = strings.TrimRight(value.Value, "/")
			}

			if value.Platform != "" {
				setValue(value.Audience, value.Platform, name, value.Value)
			}
			for _, platform := range value.Platforms {
				setValue(value.Audience, platform, name, value.Value)
			}
		}
	}
}

func updateMap[K comparable, V any](target map[K]V, updateWith map[K]V) {
	for key, value := range updateWith {
		target[key] = value
	}
}

// ExpandVariables converts "{variable name}" to the value that belongs to the given audience and platform.
func (c *Conf) ExpandVariables(valueToExpand string, audience VariableAudience, platform VariablePlatform) string {
	varsForPlatform := c.getVariables(audience, platform)
	if len(varsForPlatform) == 0 {
		log.Warn().
			Str("valueToExpand", valueToExpand).
			Str("audience", string(audience)).
			Str("platform", string(platform)).
			Msg("no variables defined for this platform given this audience")
		return valueToExpand
	}

	// Variable replacement
	for varname, varvalue := range varsForPlatform {
		placeholder := fmt.Sprintf("{%s}", varname)
		valueToExpand = strings.Replace(valueToExpand, placeholder, varvalue, -1)
	}

	// TODO: this needs to go through multiple variable replacements, to make sure
	// that, for example, the `{jobs}` variable gets the two-way variables applied
	// as well.

	return valueToExpand
}

// GetVariables returns the variable values for this (audience, platform) combination.
// If no variables are found, just returns an empty map. If a value is defined
// for both the "all" platform and specifically the given platform, the specific
// platform definition wins.
func (c *Conf) getVariables(audience VariableAudience, platform VariablePlatform) map[string]string {
	platformsForAudience := c.VariablesLookup[audience]
	if len(platformsForAudience) == 0 {
		return make(map[string]string)
	}

	varsForPlatform := map[string]string{}
	updateMap(varsForPlatform, platformsForAudience[VariablePlatformAll])
	updateMap(varsForPlatform, platformsForAudience[platform])
	return varsForPlatform
}

// ResolveVariables returns the variables for this (audience, platform) combination.
// If no variables are found, just returns an empty map. If a value is defined
// for both the "all" platform and specifically the given platform, the specific
// platform definition wins.
func (c *Conf) ResolveVariables(audience VariableAudience, platform VariablePlatform) map[string]ResolvedVariable {
	varsForPlatform := c.getVariables(audience, platform)

	resolvedVars := make(map[string]ResolvedVariable)
	for name, value := range varsForPlatform {
		resolvedVar := ResolvedVariable{Value: value}

		// Find the 'IsTwoway' property by finding the actual foundVar, which can be
		// defined in two places.
		if foundVar, ok := c.Variables[name]; ok {
			resolvedVar.IsTwoWay = foundVar.IsTwoWay
		} else if foundVar, ok := c.implicitVariables[name]; ok {
			resolvedVar.IsTwoWay = foundVar.IsTwoWay
		} else {
			log.Error().Str("variable", name).Msg("unable to find this variable, where did it come from?")
		}

		resolvedVars[name] = resolvedVar
	}

	return resolvedVars
}

// checkVariables performs some basic checks on variable definitions.
// All errors are logged, not returned.
func (c *Conf) checkVariables() {
	for name, variable := range c.Variables {
		for valueIndex, value := range variable.Values {
			// No platforms at all.
			if value.Platform == "" && len(value.Platforms) == 0 {
				log.Error().
					Str("name", name).
					Interface("value", value).
					Msg("variable has a platformless value")
				continue
			}

			// Both Platform and Platforms.
			if value.Platform != "" && len(value.Platforms) > 0 {
				log.Warn().
					Str("name", name).
					Interface("value", value).
					Str("platform", string(value.Platform)).
					Interface("platforms", value.Platforms).
					Msg("variable has a both 'platform' and 'platforms' set")
				value.Platforms = append(value.Platforms, value.Platform)
				value.Platform = ""
			}

			if value.Audience == "" {
				value.Audience = "all"
			} else if !validAudiences[value.Audience] {
				log.Error().
					Str("name", name).
					Interface("value", value).
					Str("audience", string(value.Audience)).
					Msg("variable invalid audience")
			}

			variable.Values[valueIndex] = value
		}
	}
}

func (c *Conf) checkDatabase() {
	c.DatabaseDSN = strings.TrimSpace(c.DatabaseDSN)
}

// Overwrite stores this configuration object as flamenco-manager.yaml.
func (c *Conf) Overwrite() error {
	tempFilename := configFilename + "~"
	if err := c.Write(tempFilename); err != nil {
		return fmt.Errorf("writing config to %s: %w", tempFilename, err)
	}
	if err := os.Rename(tempFilename, configFilename); err != nil {
		return fmt.Errorf("moving %s to %s: %w", tempFilename, configFilename, err)
	}

	log.Info().Str("filename", configFilename).Msg("saved configuration to file")
	return nil
}

// Write saves the current in-memory configuration to a YAML file.
func (c *Conf) Write(filename string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	fmt.Fprintf(f, "# Configuration file for %s.\n", appinfo.ApplicationName)
	fmt.Fprintln(f, "# For an explanation of the fields, refer to flamenco-manager-example.yaml")
	fmt.Fprintln(f, "#")
	fmt.Fprintln(f, "# NOTE: this file will be overwritten by Flamenco Manager's web-based configuration system.")
	fmt.Fprintln(f, "#")
	now := time.Now()
	fmt.Fprintf(f, "# This file was written on %s by %s\n\n",
		now.Format("2006-01-02 15:04:05 -07:00"),
		appinfo.FormattedApplicationInfo(),
	)

	n, err := f.Write(data)
	if err != nil {
		return err
	}
	if n < len(data) {
		return io.ErrShortWrite
	}
	if err = f.Close(); err != nil {
		return err
	}

	log.Debug().Str("filename", filename).Msg("config file written")
	return nil
}

// GetTestConfig returns the configuration for unit tests.
// The config is loaded from `test-flamenco-manager.yaml` in the directory
// containing the caller's source.
// The `overrides` parameter can be used to override configuration between
// loading it and processing the file's contents.
func GetTestConfig(overrides ...func(c *Conf)) Conf {
	_, myFilename, _, _ := runtime.Caller(1)
	myDir := path.Dir(myFilename)

	filepath := path.Join(myDir, "test-flamenco-manager.yaml")
	conf, err := loadConf(filepath, overrides...)
	if err != nil {
		log.Fatal().Err(err).Str("file", filepath).Msg("unable to load test config")
	}

	return conf
}
