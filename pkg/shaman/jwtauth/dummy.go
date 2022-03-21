package jwtauth

// SPDX-License-Identifier: GPL-3.0-or-later

/* This is just a dummy package. We still have to properly design authentication
 * for Flamenco 3, but the ported code from Flamenco 2's Shaman implementation
 * uses JWT Authentication.
 */

type Authenticator interface {
}

type AlwaysDeny struct{}

var _ Authenticator = (*AlwaysDeny)(nil)
