package checkout

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"fmt"

	"git.blender.org/flamenco/pkg/api"
	"git.blender.org/flamenco/pkg/shaman/filestore"

	"github.com/rs/zerolog"
)

func (m *Manager) ReportRequirements(ctx context.Context, requirements api.ShamanRequirementsRequest) (api.ShamanRequirementsResponse, error) {
	logger := zerolog.Ctx(ctx)
	logger.Debug().Msg("user requested checkout requirements")

	missing := api.ShamanRequirementsResponse{}
	alreadyRequested := map[string]bool{}
	for _, fileSpec := range requirements.Files {
		fileKey := fmt.Sprintf("%s/%d", fileSpec.Sha, fileSpec.Size)
		if alreadyRequested[fileKey] {
			// User asked for this (checksum, filesize) tuple already.
			continue
		}

		path, status := m.fileStore.ResolveFile(fileSpec.Sha, int64(fileSpec.Size), filestore.ResolveEverything)

		var apiStatus api.ShamanFileStatus
		switch status {
		case filestore.StatusDoesNotExist:
			// Caller can upload this file immediately.
			apiStatus = api.ShamanFileStatusUnknown
		case filestore.StatusUploading:
			// Caller should postpone uploading this file until all 'unknown' files have been uploaded.
			apiStatus = api.ShamanFileStatusUploading
		case filestore.StatusStored:
			// We expect this file to be sent soon, though, so we need to
			// 'touch' it to make sure it won't be GC'd in the mean time.
			go touchFile(path)

			// Only send a response when the caller needs to do something.
			continue
		default:
			logger.Error().
				Str("path", path).
				Str("status", status.String()).
				Str("checksum", fileSpec.Sha).
				Int("filesize", fileSpec.Size).
				Msg("invalid status returned by ResolveFile")
			continue
		}

		alreadyRequested[fileKey] = true
		missing.Files = append(missing.Files, api.ShamanFileSpecWithStatus{
			ShamanFileSpec: fileSpec,
			Status:         apiStatus,
		})
	}

	return missing, nil
}
