package config

/* ***** BEGIN GPL LICENSE BLOCK *****
 *
 * Original Code Copyright (C) 2022 Blender Foundation.
 *
 * This file is part of Flamenco.
 *
 * Flamenco is free software: you can redistribute it and/or modify it under
 * the terms of the GNU General Public License as published by the Free Software
 * Foundation, either version 3 of the License, or (at your option) any later
 * version.
 *
 * Flamenco is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR
 * A PARTICULAR PURPOSE.  See the GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * Flamenco.  If not, see <https://www.gnu.org/licenses/>.
 *
 * ***** END GPL LICENSE BLOCK ***** */

import (
	"time"

	"git.blender.org/flamenco/pkg/api"
)

// The default configuration, use DefaultConfig() to obtain a copy.
var defaultConfig = Conf{
	Base: Base{
		Meta: ConfMeta{Version: latestConfigVersion},

		ManagerName:  "Flamenco Manager",
		Listen:       ":8080",
		ListenHTTPS:  ":8433",
		DatabaseDSN:  "flamenco-manager.sqlite",
		TaskLogsPath: "./task-logs",
		// DownloadTaskSleep:           10 * time.Minute,
		// DownloadTaskRecheckThrottle: 10 * time.Second,
		// TaskUpdatePushMaxInterval:   5 * time.Second,
		// TaskUpdatePushMaxCount:      3000,
		// CancelTaskFetchInterval:     10 * time.Second,
		ActiveTaskTimeoutInterval:   10 * time.Minute,
		ActiveWorkerTimeoutInterval: 1 * time.Minute,
		// FlamencoStr:                 defaultServerURL,

		// // Days are assumed to be 24 hours long. This is not exactly accurate, but should
		// // be accurate enough for this type of cleanup.
		// TaskCleanupMaxAge: 14 * 24 * time.Hour,
		SSDPDiscovery: false, // Only enable after SSDP discovery has been improved (avoid finding printers).

		BlacklistThreshold:         3,
		TaskFailAfterSoftFailCount: 3,

		WorkerCleanupStatus: []string{string(api.WorkerStatusOffline)},

		TestTasks: TestTasks{
			BlenderRender: BlenderRenderConfig{
				JobStorage:   "{job_storage}/test-jobs",
				RenderOutput: "{render}/test-renders",
			},
		},

		Shaman: ShamanConfig{
			Enabled:       true,
			FileStorePath: defaultShamanFilestorePath,
			GarbageCollect: ShamanGarbageCollect{
				Period:            24 * time.Hour,
				MaxAge:            31 * 24 * time.Hour,
				ExtraCheckoutDirs: []string{},
			},
		},

		// JWT: jwtauth.Config{
		// 	DownloadKeysInterval: 1 * time.Hour,
		// },
	},

	Variables: map[string]Variable{
		// The default commands assume that the executables are available on $PATH.
		"blender": {
			Direction: "oneway",
			Values: VariableValues{
				VariableValue{Platform: "linux", Value: "blender --factory-startup --background"},
				VariableValue{Platform: "windows", Value: "blender.exe --factory-startup --background"},
				VariableValue{Platform: "darwin", Value: "blender --factory-startup --background"},
			},
		},
		"ffmpeg": {
			Direction: "oneway",
			Values: VariableValues{
				VariableValue{Platform: "linux", Value: "ffmpeg"},
				VariableValue{Platform: "windows", Value: "ffmpeg.exe"},
				VariableValue{Platform: "darwin", Value: "ffmpeg"},
			},
		},
		// TODO: determine useful defaults for these.
		"job_storage": {
			Direction: "twoway",
			Values: VariableValues{
				VariableValue{Platform: "linux", Value: "/shared/flamenco/jobs"},
				VariableValue{Platform: "windows", Value: "S:/flamenco/jobs"},
				VariableValue{Platform: "darwin", Value: "/Volumes/Shared/flamenco/jobs"},
			},
		},
		"render": {
			Direction: "twoway",
			Values: VariableValues{
				VariableValue{Platform: "linux", Value: "/shared/flamenco/render"},
				VariableValue{Platform: "windows", Value: "S:/flamenco/render"},
				VariableValue{Platform: "darwin", Value: "/Volumes/Shared/flamenco/render"},
			},
		},
	},
}
