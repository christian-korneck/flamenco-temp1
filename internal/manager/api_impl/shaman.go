package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"git.blender.org/flamenco/pkg/api"
	"git.blender.org/flamenco/pkg/shaman/fileserver"
)

func (f *Flamenco) isShamanEnabled() bool {
	return f.shaman.IsEnabled()
}

// Create a directory, and symlink the required files into it. The files must all have been uploaded to Shaman before calling this endpoint.
// (POST /shaman/checkout/create/{checkoutID})
func (f *Flamenco) ShamanCheckout(e echo.Context) error {
	logger := requestLogger(e)
	if !f.isShamanEnabled() {
		logger.Error().Msg("shaman server not active, unable to serve request")
		return sendAPIError(e, http.StatusServiceUnavailable, "shaman server not active")
	}

	var reqBody api.ShamanCheckoutJSONBody
	err := e.Bind(&reqBody)
	if err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}

	err = f.shaman.Checkout(e.Request().Context(), api.ShamanCheckout(reqBody))
	if err != nil {
		// TODO: return 409 when checkout already exists.
		logger.Warn().Err(err).Msg("Shaman: creating checkout")
		return sendAPIError(e, http.StatusInternalServerError, "unexpected error: %v", err)
	}

	return e.String(http.StatusNoContent, "")
}

// Checks a Shaman Requirements file, and reports which files are unknown.
// (POST /shaman/checkout/requirements)
func (f *Flamenco) ShamanCheckoutRequirements(e echo.Context) error {
	logger := requestLogger(e)
	if !f.isShamanEnabled() {
		logger.Error().Msg("shaman server not active, unable to serve request")
		return sendAPIError(e, http.StatusServiceUnavailable, "shaman server not active")
	}

	var reqBody api.ShamanCheckoutRequirementsJSONBody
	err := e.Bind(&reqBody)
	if err != nil {
		logger.Warn().Err(err).Msg("bad request received")
		return sendAPIError(e, http.StatusBadRequest, "invalid format")
	}

	unknownFiles, err := f.shaman.Requirements(e.Request().Context(), api.ShamanRequirementsRequest(reqBody))
	if err != nil {
		logger.Warn().Err(err).Msg("Shaman: checking checkout requirements file")
		return sendAPIError(e, http.StatusInternalServerError, "unexpected error: %v", err)
	}

	return e.JSON(http.StatusOK, unknownFiles)
}

// Check the status of a file on the Shaman server.
// (OPTIONS /shaman/files/{checksum}/{filesize})
func (f *Flamenco) ShamanFileStoreCheck(e echo.Context, checksum string, filesize int) error {
	logger := requestLogger(e).With().
		Str("checksum", checksum).Int("filesize", filesize).
		Logger()
	if !f.isShamanEnabled() {
		logger.Error().Msg("shaman server not active, unable to serve request")
		return sendAPIError(e, http.StatusServiceUnavailable, "shaman server not active")
	}

	logger.Debug().Msg("shaman: checking file")
	status := f.shaman.FileStoreCheck(e.Request().Context(), checksum, int64(filesize))

	return e.JSON(http.StatusOK, api.ShamanSingleFileStatus{
		Status: status,
	})
}

// Store a new file on the Shaman server. Note that the Shaman server can
// forcibly close the HTTP connection when another client finishes uploading the
// exact same file, to prevent double uploads.
// (POST /shaman/files/{checksum}/{filesize})
func (f *Flamenco) ShamanFileStore(e echo.Context, checksum string, filesize int, params api.ShamanFileStoreParams) error {
	var (
		origFilename string
		canDefer     bool
	)

	logCtx := requestLogger(e).With().
		Str("checksum", checksum).Int("filesize", filesize)

	if !f.isShamanEnabled() {
		logger := logCtx.Logger()
		logger.Error().Msg("shaman server not active, unable to serve request")
		return sendAPIError(e, http.StatusServiceUnavailable, "shaman server not active")
	}

	if params.XShamanCanDeferUpload != nil {
		canDefer = *params.XShamanCanDeferUpload
		logCtx = logCtx.Bool("canDefer", canDefer)
	}
	if params.XShamanOriginalFilename != nil {
		origFilename = *params.XShamanOriginalFilename
		logCtx = logCtx.Str("originalFilename", origFilename)
	}
	logger := logCtx.Logger()

	err := f.shaman.FileStore(e.Request().Context(), e.Request().Body,
		checksum, int64(filesize),
		canDefer, origFilename,
	)
	if err != nil {
		switch err {
		case fileserver.ErrFileAlreadyExists:
			return e.String(http.StatusAlreadyReported, "")
		case fileserver.ErrFileShouldDefer:
			return e.String(http.StatusTooEarly, "")
		}

		logger.Warn().Err(err).Msg("shaman: checking stored file")

		switch v := err.(type) {
		case fileserver.ErrFileSizeMismatch, fileserver.ErrFileChecksumMismatch:
			return sendAPIError(e, http.StatusExpectationFailed, v.Error())
		default:
			return sendAPIError(e, http.StatusInternalServerError, "unexpected error: %v", err)
		}
	}

	return nil
}
