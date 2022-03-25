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

	missing := api.ShamanRequirementsResponse{
		Files: []api.ShamanFileSpecWithStatus{},
	}

	alreadyRequested := map[string]bool{}
	for _, fileSpec := range requirements.Files {
		fileKey := fmt.Sprintf("%s/%d", fileSpec.Sha, fileSpec.Size)
		if alreadyRequested[fileKey] {
			// User asked for this (checksum, filesize) tuple already.
			continue
		}

		storePath, status := m.fileStore.ResolveFile(fileSpec.Sha, int64(fileSpec.Size), filestore.ResolveEverything)

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
			go func() {
				if err := touchFile(storePath); err != nil {
					logger.Error().Err(err).Str("path", storePath).Msg("shaman: error touching file")
				}
			}()
			// Only send a response when the caller needs to do something.
			continue
		default:
			logger.Error().
				Str("path", fileSpec.Path).
				Str("status", status.String()).
				Str("checksum", fileSpec.Sha).
				Int("filesize", fileSpec.Size).
				Msg("shaman: invalid status returned by ResolveFile, ignoring this file")
			continue
		}

		alreadyRequested[fileKey] = true
		fileSpec := api.ShamanFileSpecWithStatus{
			Path:   fileSpec.Path,
			Sha:    fileSpec.Sha,
			Size:   fileSpec.Size,
			Status: apiStatus,
		}
		logger.Trace().Interface("fileSpec", fileSpec).Msg("shaman: file needed from client")
		missing.Files = append(missing.Files, fileSpec)
	}

	return missing, nil
}
