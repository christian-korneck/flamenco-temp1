package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"git.blender.org/flamenco/internal/manager/config"
	"git.blender.org/flamenco/pkg/api"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
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

func TestCheckSharedStoragePath(t *testing.T) {
	mf, finish := metaTestFixtures(t)
	defer finish()

	doTest := func(path string) echo.Context {
		echoCtx := mf.prepareMockedJSONRequest(
			api.PathCheckInput{Path: path})
		err := mf.flamenco.CheckSharedStoragePath(echoCtx)
		if !assert.NoError(t, err) {
			t.FailNow()
		}
		return echoCtx
	}

	// Test empty path.
	echoCtx := doTest("")
	assertResponseJSON(t, echoCtx, http.StatusOK, api.PathCheckResult{
		Path:     "",
		IsUsable: false,
		Cause:    "An empty path is never suitable as shared storage",
	})

	// Test usable path (well, at least readable & writable; it may not be shared via Samba/NFS).
	echoCtx = doTest(mf.tempdir)
	assertResponseJSON(t, echoCtx, http.StatusOK, api.PathCheckResult{
		Path:     mf.tempdir,
		IsUsable: true,
		Cause:    "Directory checked OK!",
	})
	files, err := filepath.Glob(filepath.Join(mf.tempdir, "*"))
	if assert.NoError(t, err) {
		assert.Empty(t, files, "After a query, there should not be any leftovers")
	}

	// Test inaccessible path.
	{
		parentPath := filepath.Join(mf.tempdir, "deep")
		testPath := filepath.Join(parentPath, "nesting")
		if err := os.Mkdir(parentPath, fs.ModePerm); !assert.NoError(t, err) {
			t.FailNow()
		}
		if err := os.Mkdir(testPath, fs.FileMode(0)); !assert.NoError(t, err) {
			t.FailNow()
		}
		echoCtx := doTest(testPath)
		result := api.PathCheckResult{}
		getResponseJSON(t, echoCtx, http.StatusOK, &result)
		assert.Equal(t, testPath, result.Path)
		assert.False(t, result.IsUsable)
		assert.Contains(t, result.Cause, "Unable to create a file")
	}
}

func metaTestFixtures(t *testing.T) (mockedFlamenco, func()) {
	mockCtrl := gomock.NewController(t)
	mf := newMockedFlamenco(mockCtrl)

	tempdir, err := os.MkdirTemp("", "test-temp-dir")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	mf.tempdir = tempdir

	finish := func() {
		mockCtrl.Finish()
		os.RemoveAll(tempdir)
	}

	return mf, finish
}
