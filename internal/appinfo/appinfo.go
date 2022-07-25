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

// FormattedApplicationInfo returns the application name & version as single string.
func FormattedApplicationInfo() string {
	return fmt.Sprintf("%s %s", ApplicationName, ApplicationVersion)
}

// UserAgent returns the application name & version suitable for the HTTP User-Agent header.
func UserAgent() string {
	return fmt.Sprintf("%s/%s (%s)", ApplicationName, ApplicationVersion, ApplicationGitHash)
}
