package main

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
)

const addonVersionFile = "addon/flamenco/__init__.py"

// updateAddon changes the version number in the Blender add-on.
// Returns whether the file actually changed.
func updateAddon() bool {
	// The add-on needs a (x, y, z) tuple as version, and doesn't support suffixes
	// like `-dev0` or `-beta3`.
	splitOnDash := strings.SplitN(cliArgs.newVersion, "-", 2)
	versionWithoutSuffix := splitOnDash[0]

	var warning string
	if len(splitOnDash) >= 2 {
		log.Warn().Msg("versions of form `x.y.z-something` will put a warning in the bl_info about the `-something`")
		warning = fmt.Sprintf("This is version %s of the add-on, which is not a stable release", cliArgs.newVersion)
	}

	versionParts := strings.Split(versionWithoutSuffix, ".")
	var versionTuple string
	switch len(versionParts) {
	case 0:
		log.Fatal().Str("versionWithoutSuffix", versionWithoutSuffix).Msg("no dot-separated version number found")
	case 1:
		log.Warn().Strs("versionParts", versionParts).Msg("only a major version found, may be the wrong syntax")
		versionTuple = fmt.Sprintf("(%s, 0)", versionParts[0])
	case 2:
		log.Debug().Strs("versionParts", versionParts).Msg("major.minor version found, this is expected")
		versionTuple = fmt.Sprintf("(%s, %s)", versionParts[0], versionParts[1])
	case 3:
		log.Debug().Strs("versionParts", versionParts).Msg("major.minor.micro version found, this is expected")
		versionTuple = fmt.Sprintf("(%s, %s, %s)", versionParts[0], versionParts[1], versionParts[2])
	default:
		log.Warn().Strs("versionParts", versionParts).Msg("more than three (major, minor, micro) version parts found, using only the first three")
		versionTuple = fmt.Sprintf("(%s, %s, %s)", versionParts[0], versionParts[1], versionParts[2])
	}

	var blinfoOpened, blinfoClosed bool
	replacer := func(line string) string {
		switch {
		case !blinfoOpened && strings.HasPrefix(line, "bl_info = {"):
			blinfoOpened = true
		case blinfoOpened && strings.HasPrefix(line, "}"):
			blinfoClosed = true
		case blinfoOpened && !blinfoClosed && strings.HasPrefix(line, "    \"version\":"):
			return fmt.Sprintf("    \"version\": %s,", versionTuple)
		case blinfoOpened && !blinfoClosed && strings.HasPrefix(line, "    \"warning\":"):
			return fmt.Sprintf("    \"warning\": %q,", warning)
		}
		return line
	}

	fileWasChanged, err := updateLines(addonVersionFile, replacer)
	if err != nil {
		log.Fatal().Err(err).Msg("error updating add-on")
	}
	return fileWasChanged
}
