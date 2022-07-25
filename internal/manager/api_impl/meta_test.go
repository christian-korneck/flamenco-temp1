package api_impl

// SPDX-License-Identifier: GPL-3.0-or-later

import (
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
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
		Cause:    "An empty path is not suitable as shared storage",
	})

	// Test usable path (well, at least readable & writable; it may not be shared via Samba/NFS).
	echoCtx = doTest(mf.tempdir)
	assertResponseJSON(t, echoCtx, http.StatusOK, api.PathCheckResult{
		Path:     mf.tempdir,
		IsUsable: true,
		Cause:    "Directory checked successfully",
	})
	files, err := filepath.Glob(filepath.Join(mf.tempdir, "*"))
	if assert.NoError(t, err) {
		assert.Empty(t, files, "After a query, there should not be any leftovers")
	}

	// Test inaccessible path.
	// For some reason, this doesn't work on Windows, and creating a file in
	// that directory is still allowed. The Explorer's properties panel of the
	// directory also shows "Read Only (only applies to files)", so at least
	// that seems consistent.
	// FIXME: find another way to test with unwritable directories on Windows.
	if runtime.GOOS != "windows" {
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

func TestSaveSetupAssistantConfig(t *testing.T) {
	mf, finish := metaTestFixtures(t)
	defer finish()

	doTest := func(body api.SetupAssistantConfig) config.Conf {
		// Always start the test with a clean configuration.
		originalConfig := config.DefaultConfig(func(c *config.Conf) {
			c.SharedStoragePath = ""
		})
		var savedConfig config.Conf

		// Mock the loading & saving of the config.
		mf.config.EXPECT().Get().Return(&originalConfig)
		mf.config.EXPECT().Save().Do(func() error {
			savedConfig = originalConfig
			return nil
		})

		// Call the API.
		echoCtx := mf.prepareMockedJSONRequest(body)
		err := mf.flamenco.SaveSetupAssistantConfig(echoCtx)
		if !assert.NoError(t, err) {
			t.FailNow()
		}

		assertResponseNoContent(t, echoCtx)
		return savedConfig
	}

	// Test situation where file association with .blend files resulted in a blender executable.
	{
		savedConfig := doTest(api.SetupAssistantConfig{
			StorageLocation: mf.tempdir,
			BlenderExecutable: api.BlenderPathCheckResult{
				IsUsable: true,
				Input:    "",
				Path:     "/path/to/blender",
				Source:   api.BlenderPathSourceFileAssociation,
			},
		})
		assert.Equal(t, mf.tempdir, savedConfig.SharedStoragePath)
		expectBlenderVar := config.Variable{
			Values: config.VariableValues{
				{Platform: "linux", Value: "blender " + config.DefaultBlenderArguments},
				{Platform: "windows", Value: "blender " + config.DefaultBlenderArguments},
				{Platform: "darwin", Value: "blender " + config.DefaultBlenderArguments},
			},
		}
		assert.Equal(t, expectBlenderVar, savedConfig.Variables["blender"])
	}

	// Test situation where the given command could be found on $PATH.
	{
		savedConfig := doTest(api.SetupAssistantConfig{
			StorageLocation: mf.tempdir,
			BlenderExecutable: api.BlenderPathCheckResult{
				IsUsable: true,
				Input:    "kitty",
				Path:     "/path/to/kitty",
				Source:   api.BlenderPathSourcePathEnvvar,
			},
		})
		assert.Equal(t, mf.tempdir, savedConfig.SharedStoragePath)
		expectBlenderVar := config.Variable{
			Values: config.VariableValues{
				{Platform: "linux", Value: "kitty " + config.DefaultBlenderArguments},
				{Platform: "windows", Value: "kitty " + config.DefaultBlenderArguments},
				{Platform: "darwin", Value: "kitty " + config.DefaultBlenderArguments},
			},
		}
		assert.Equal(t, expectBlenderVar, savedConfig.Variables["blender"])
	}

	// Test a custom command given with the full path.
	{
		savedConfig := doTest(api.SetupAssistantConfig{
			StorageLocation: mf.tempdir,
			BlenderExecutable: api.BlenderPathCheckResult{
				IsUsable: true,
				Input:    "/bin/cat",
				Path:     "/bin/cat",
				Source:   api.BlenderPathSourceInputPath,
			},
		})
		assert.Equal(t, mf.tempdir, savedConfig.SharedStoragePath)
		expectBlenderVar := config.Variable{
			Values: config.VariableValues{
				{Platform: "linux", Value: "/bin/cat " + config.DefaultBlenderArguments},
				{Platform: "windows", Value: "/bin/cat " + config.DefaultBlenderArguments},
				{Platform: "darwin", Value: "/bin/cat " + config.DefaultBlenderArguments},
			},
		}
		assert.Equal(t, expectBlenderVar, savedConfig.Variables["blender"])
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
