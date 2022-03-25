package checkout

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"git.blender.org/flamenco/pkg/api"
	"git.blender.org/flamenco/pkg/shaman/filestore"
	"github.com/rs/zerolog"
)

var (
	ErrMissingFiles = errors.New("unknown files requested in checkout")

	validCheckoutRegexp = regexp.MustCompile(`^[^/?*:;{}\\][^?*:;{}\\]*$`)
)

func (m *Manager) Checkout(ctx context.Context, checkout api.ShamanCheckout) error {
	logger := (*zerolog.Ctx(ctx)).With().
		Str("checkoutPath", checkout.CheckoutPath).Logger()
	logger.Debug().Msg("shaman: user requested checkout creation")

	// Actually create the checkout.
	resolvedCheckoutInfo, err := m.PrepareCheckout(checkout.CheckoutPath)
	if err != nil {
		return err
	}

	// The checkout directory was created, so if anything fails now, it should be erased.
	var checkoutOK bool
	defer func() {
		if !checkoutOK {
			err := m.EraseCheckout(checkout.CheckoutPath)
			if err != nil {
				logger.Error().Err(err).Msg("shaman: error erasing checkout directory")
			}
		}
	}()

	for _, fileSpec := range checkout.Files {
		blobPath, status := m.fileStore.ResolveFile(fileSpec.Sha, int64(fileSpec.Size), filestore.ResolveStoredOnly)
		if status != filestore.StatusStored {
			// Caller should upload this file before we can create the checkout.
			return ErrMissingFiles
		}

		if err := m.SymlinkToCheckout(blobPath, resolvedCheckoutInfo.absolutePath, fileSpec.Path); err != nil {
			return fmt.Errorf("symlinking %q to checkout: %w", fileSpec.Path, err)
		}
	}

	checkoutOK = true // Prevent the checkout directory from being erased again.
	logger.Info().Msg("shaman: checkout created")
	return nil
}

func isValidCheckoutPath(checkoutPath string) bool {
	if !validCheckoutRegexp.MatchString(checkoutPath) {
		return false
	}
	if strings.Contains(checkoutPath, "../") || strings.Contains(checkoutPath, "/..") {
		return false
	}
	return true
}
