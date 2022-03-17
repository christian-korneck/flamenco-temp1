# Flamenco-NG Features

This file lists the features that ought to be implemented at some point in time.
It will likely be expanded over time while forgotten features of Flamenco v2 are
rediscovered.

Note that list is **not** in any specific order.

## Manager

- [x] OpenAPI implementation that actually runs as a server process
- [x] Process lifecycle management, clean shutdown
- [x] Task Scheduler
- [x] Unify command definition (type/parameters in persistence layer, name/settings in API layer)
- [x] Job status change handling
- [ ] Port the old 'fail-requested' task status handling code to the new Manager
- [ ] Task timeout monitoring
- [ ] Worker Blacklisting & failed task requeueing
- [ ] Worker timeout monitoring
- [x] Worker-assigned task re-queueing on sign-off
- [x] Variable replacement
- [x] Configuration file reading
- [ ] Configuration provisioning support
- [ ] Shaman integration
- [ ] Worker sleep schedule
- [ ] Web frontend for Worker management
- [ ] Web frontend for Job & Task management
- [ ] Web frontend for configuration management
- [ ] Frontend authentication
- [x] Worker API authentication
- [ ] User/Job Submission API authentication
- [ ] Last rendered image display
- [ ] Per-job last rendered image
- [ ] Auto-removal of old Workers
- [ ] Shaman/Job settings: Distinction between "can be auto-deleted" jobs and "must be kept around" jobs.
- [ ] Sending data upstream to Flamenco Overwatch
- [ ] Job archival / task cleanup

## Worker

- [x] OpenAPI client that actually runs as a process
- [x] Process lifecycle management, clean shutdown
- [x] Task executor
- [x] Command runner
- [x] Log uploader
- [ ] POSIX signal handling (sleep/wake up)
- [x] Implement `frames-to-video` command.
- [x] Implement `move-directory` command.
- [ ] Refactor CLI-running code by analyzing the current Blender and FFmpeg commands.

## Both

- [x] Worker registration & auto-reregistration
- [x] Worker API authentication
- [x] Worker sign on/off
- [ ] Worker status change management, including lazy status changes
- [x] Task status change handling
- [ ] Worker test mode
- [ ] Worker registration secret (to protect the Manager from arbitrary Workers registering)
- [x] UPnP/SSDP for automatic Manager finding

## Blender Add-on

- [x] Communication with the Flamenco Manager API
- [ ] Job construction
- [x] BAT-packing
- [ ] uploading to Shaman
- [ ] User authentication
- [ ] Editable job names (for rendering multiple jobs from the same filename)
- [ ] Add `options={'HIDDEN'}` to the Flamenco RNA properties.
- [ ] Add eTag-like checksum to the job types. The aim is to prevent artists from sending in a new job with an old job. This can happen when the job type is changed on the Manager, but in Blender the "refresh" button wasn't pressed yet. This should only cover the job settings, not the rest of the compiler script.
- [ ] Frame Range: allow automatically syncing with scene frame start/end.
- [ ] Don't take output settings from the current scene, but have separate settings for Flamenco. To be fleshed out still.

## Nice to have / new features

- [ ] Notification system to push "job done" messages to. Ideally would be in a form/shape that allows sending a message to Rocket.Chat, Matrix, Telegram, Discord, email, webbrowser, push URL-encoded/JSON/XML to some URL, stuff like that. Idea by Dan McLaughlin.
- [ ] Notification client inside Blender itself, so that you get a message when your job is done.
