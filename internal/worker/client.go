package worker

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"git.blender.org/flamenco/pkg/api"
)

// Generate the mock for the client interface.
//go:generate go run github.com/golang/mock/mockgen -destination mocks/client.gen.go -package mocks git.blender.org/flamenco/internal/worker FlamencoClient

// FlamencoClient is a wrapper for api.ClientWithResponsesInterface so that locally mocks can be created.
type FlamencoClient interface {
	api.ClientWithResponsesInterface
}
