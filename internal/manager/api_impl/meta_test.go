package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"net/http"
	"testing"

	"git.blender.org/flamenco/internal/manager/config"
	"git.blender.org/flamenco/pkg/api"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetVariables(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mf := newMockedFlamenco(mockCtrl)

	// Test Linux Worker.
	{
		resolvedVarsLinuxWorker := make(map[string]config.ResolvedVariable)
		resolvedVarsLinuxWorker["jobs"] = config.ResolvedVariable{
			IsTwoWay: true,
			Value:    "Linux value",
		}
		resolvedVarsLinuxWorker["blender"] = config.ResolvedVariable{
			IsTwoWay: false,
			Value:    "/usr/local/blender",
		}

		mf.config.EXPECT().
			ResolveVariables(config.VariableAudienceWorkers, config.VariablePlatformLinux).
			Return(resolvedVarsLinuxWorker)

		echoCtx := mf.prepareMockedRequest(nil)
		err := mf.flamenco.GetVariables(echoCtx, api.ManagerVariableAudienceWorkers, "linux")
		assert.NoError(t, err)
		assertResponseJSON(t, echoCtx, http.StatusOK, api.ManagerVariables{
			AdditionalProperties: map[string]api.ManagerVariable{
				"blender": {Value: "/usr/local/blender", IsTwoway: false},
				"jobs":    {Value: "Linux value", IsTwoway: true},
			},
		})
	}

	// Test unknown platform User.
	{
		resolvedVarsUnknownPlatform := make(map[string]config.ResolvedVariable)
		mf.config.EXPECT().
			ResolveVariables(config.VariableAudienceUsers, config.VariablePlatform("troll")).
			Return(resolvedVarsUnknownPlatform)

		echoCtx := mf.prepareMockedRequest(nil)
		err := mf.flamenco.GetVariables(echoCtx, api.ManagerVariableAudienceUsers, "troll")
		assert.NoError(t, err)
		assertResponseJSON(t, echoCtx, http.StatusOK, api.ManagerVariables{})
	}
}
