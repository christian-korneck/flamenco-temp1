package config

import (
	"runtime"
	"time"

	shaman_config "git.blender.org/flamenco/pkg/shaman/config"
)

// SPDX-License-Identifier: GPL-3.0-or-later

// The default configuration, use DefaultConfig() to obtain a copy.
var defaultConfig = Conf{
	Base: Base{
		Meta: ConfMeta{Version: latestConfigVersion},

		ManagerName: "Flamenco Manager",
		Listen:      ":8080",
		// ListenHTTPS:   ":8433",
		DatabaseDSN:   "flamenco-manager.sqlite",
		TaskLogsPath:  "./task-logs",
		SSDPDiscovery: true,
		StoragePath:   "./flamenco-storage",

		Shaman: shaman_config.Config{
			// Enable Shaman by default, except on Windows where symlinks are still tricky.
			Enabled: runtime.GOOS != "windows",
			GarbageCollect: shaman_config.GarbageCollect{
				Period:            24 * time.Hour,
				MaxAge:            31 * 24 * time.Hour,
				ExtraCheckoutDirs: []string{},
			},
		},

		TaskTimeout: 10 * time.Minute,
		// ActiveWorkerTimeoutInterval: 1 * time.Minute,

		// // Days are assumed to be 24 hours long. This is not exactly accurate, but should
		// // be accurate enough for this type of cleanup.
		// TaskCleanupMaxAge: 14 * 24 * time.Hour,

		// BlacklistThreshold:         3,
		// TaskFailAfterSoftFailCount: 3,

		// WorkerCleanupStatus: []string{string(api.WorkerStatusOffline)},

		// TestTasks: TestTasks{
		// 	BlenderRender: BlenderRenderConfig{
		// 		JobStorage:   "{job_storage}/test-jobs",
		// 		RenderOutput: "{render}/test-renders",
		// 	},
		// },

		// JWT: jwtauth.Config{
		// 	DownloadKeysInterval: 1 * time.Hour,
		// },
	},

	Variables: map[string]Variable{
		// The default commands assume that the executables are available on $PATH.
		"blender": {
			Values: VariableValues{
				VariableValue{Platform: "linux", Value: "blender --factory-startup --background"},
				VariableValue{Platform: "windows", Value: "blender.exe --factory-startup --background"},
				VariableValue{Platform: "darwin", Value: "blender --factory-startup --background"},
			},
		},
		"ffmpeg": {
			Values: VariableValues{
				VariableValue{Platform: "linux", Value: "ffmpeg"},
				VariableValue{Platform: "windows", Value: "ffmpeg.exe"},
				VariableValue{Platform: "darwin", Value: "ffmpeg"},
			},
		},
		// TODO: determine useful defaults for these.
		// "job_storage": {
		// 	IsTwoWay: true,
		// 	Values: VariableValues{
		// 		VariableValue{Platform: "linux", Value: "/shared/flamenco/jobs"},
		// 		VariableValue{Platform: "windows", Value: "S:/flamenco/jobs"},
		// 		VariableValue{Platform: "darwin", Value: "/Volumes/Shared/flamenco/jobs"},
		// 	},
		// },
		// "render": {
		// 	IsTwoWay: true,
		// 	Values: VariableValues{
		// 		VariableValue{Platform: "linux", Value: "/shared/flamenco/render"},
		// 		VariableValue{Platform: "windows", Value: "S:/flamenco/render"},
		// 		VariableValue{Platform: "darwin", Value: "/Volumes/Shared/flamenco/render"},
		// 	},
		// },
	},
}
