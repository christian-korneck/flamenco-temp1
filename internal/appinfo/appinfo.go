package appinfo

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

import "fmt"

// ApplicationName contains the application name.
const ApplicationName = "Flamenco 3"

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
