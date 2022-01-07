package appinfo

import "fmt"

// ApplicationName contains the application name.
const ApplicationName = "Flamenco-NG-PoC"

// ApplicationVersion has the version number, and is set during the build.
var ApplicationVersion = "set-during-build"

// FormattedApplicationInfo returns the application name & version as single string.
func FormattedApplicationInfo() string {
	return fmt.Sprintf("%s %s", ApplicationName, ApplicationVersion)
}

// UserAgent returns the application name & version suitable for the HTTP User-Agent header.
func UserAgent() string {
	return fmt.Sprintf("%s/%s", ApplicationName, ApplicationVersion)
}
