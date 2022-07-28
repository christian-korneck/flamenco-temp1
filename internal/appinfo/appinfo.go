package appinfo

// SPDX-License-Identifier: GPL-3.0-or-later

import "fmt"

// ApplicationName contains the application name.
const ApplicationName = "Flamenco"

// ApplicationVersion is the version number of the application.
// It is set during the build.
var ApplicationVersion = "set-during-build"

// ApplicationGitHash has the Git hash of the commit used to create this build.
// It is set during the build.
var ApplicationGitHash = "set-during-build"

// ReleaseCycle determines whether this is marked as release or whether it's
// an development version. This is used to determine wehtehr ExtendedVersion()
// actually returns just the version ("release") or also has the Git hash
// (any other string).
//
// This is a string and not a boolean, because it must be set by the linker and
// that only supports strings.
var ReleaseCycle string = "set-during-build"

const releaseCycleRelease = "release"

// FormattedApplicationInfo returns the application name & version as single string.
func FormattedApplicationInfo() string {
	return fmt.Sprintf("%s %s", ApplicationName, ApplicationVersion)
}

// UserAgent returns the application name & version suitable for the HTTP User-Agent header.
func UserAgent() string {
	return fmt.Sprintf("%s/%s (%s)", ApplicationName, ApplicationVersion, ApplicationGitHash)
}

// ExtendedVersion returns the application version, and includes the Git hash if
// this is not a release version. See `IsReleaseVersion`.
func ExtendedVersion() string {
	if ReleaseCycle == releaseCycleRelease {
		return ApplicationVersion
	}

	return fmt.Sprintf("%s-%s", ApplicationVersion, ApplicationGitHash)
}
