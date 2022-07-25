package dummy

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"context"
	"errors"
	"io"

	"git.blender.org/flamenco/internal/manager/api_impl"
	"git.blender.org/flamenco/pkg/api"
)

// DummyShaman implements the Shaman interface from `internal/manager/api_impl/interfaces.go`
type DummyShaman struct{}

var _ api_impl.Shaman = (*DummyShaman)(nil)

var ErrDummyShaman = errors.New("Shaman storage component is inactive, configure Flamenco first")

func (ds *DummyShaman) IsEnabled() bool {
	return false
}
func (ds *DummyShaman) Checkout(ctx context.Context, checkout api.ShamanCheckout) (string, error) {
	return "", ErrDummyShaman
}
func (ds *DummyShaman) Requirements(ctx context.Context, requirements api.ShamanRequirementsRequest) (api.ShamanRequirementsResponse, error) {
	return api.ShamanRequirementsResponse{}, ErrDummyShaman
}
func (ds *DummyShaman) FileStoreCheck(ctx context.Context, checksum string, filesize int64) api.ShamanFileStatus {
	return api.ShamanFileStatusUnknown
}
func (ds *DummyShaman) FileStore(ctx context.Context, file io.ReadCloser, checksum string, filesize int64, canDefer bool, originalFilename string) error {
	return ErrDummyShaman
}
