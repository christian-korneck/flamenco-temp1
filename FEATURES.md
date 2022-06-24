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
- [x] Worker-assigned task re-queueing on sign-off
- [x] Variable replacement
- [x] Configuration file reading
- [x] Shaman integration
- [x] Worker API authentication
- [ ] Web frontend for Job & Task management
  - [x] Jobs list
  - [x] Job details
  - [x] Job action buttons
  - [x] Notification/status bar system
  - [x] Tasks list
  - [x] Task details
  - [x] Task action buttons
  - [x] Notification history pop-over
  - [x] Task log viewer
  - [x] Subscription system for "job detail+task" updates.
  - [x] Subscription system for "task log" updates.
  - [x] Filtering jobs by status
  - [x] Filtering tasks by status
- [ ] Make tables respect the window height
  - [ ] Jobs
  - [x] Tasks
  - [ ] Workers
- [x] Web frontend for Worker management
  - [x] "Tabs" system + Vue Router support.
  - [x] API endpoint for worker status change requests
  - [x] Web: Worker action buttons
  - [x] Implementation of lazy vs. forced status change requests
- [x] Port the old 'fail-requested' task status handling code to the new Manager
- [ ] At startup check & fix "stuck" jobs.
    - [x] Jobs in transitional statuses `cancel-requested`, `requeueing`, etc.
    - [ ] Jobs with impossible to execute tasks. For example, consider the scenario where all but one worker were blocklisted for a certain task type, and the last worker that could run it, failed it. Now if that failure was that Worker's first one, it wouldn't get blocklisted and still counts as "can execute this task type on this job". However, since it failed the task, it won't be allowed to retry it, and thus the task will get stuck in `soft-failed` status.
- [x] Task timeout monitoring
- [ ] Worker blocklisting & failed task requeueing
  - [x] Keep track of which worker failed which task.
  - [x] Clear task failure list when the task gets (re)queued.
  - [x] Keep track of a blocklist as `(worker ID, job ID, task type)` tuple in the database.
  - [ ] Show blocklist in web interface, and allow removal of workers.
- [x] Worker timeout monitoring
- [ ] Last rendered image display
  - [ ] Expand the processing queue from a single channel to a queue per job, so that a spammy job doesn't starve the other jobs from queueing images.

- [ ] Web interface: Support actions on multiple selected things
  - [ ] Workers
  - [ ] Jobs
  - [ ] Tasks

- [x] Let Manager write to task log when it's assigned to a worker.
- [ ] Worker sleep schedule
- [x] Loading of job compiler scripts from disk
- [ ] CLI option to write built-in job compiler scripts to disk
- [ ] Per-job last rendered image
- [ ] Support pausing jobs.
- [ ] Shaman/Job settings: Distinction between "can be auto-deleted" jobs and "must be kept around" jobs.
- [ ] Job archival/deletion / task cleanup
- [ ] Caching of job compilers. Currently every operation (including "fetch job
  type info", which is done by the web interface whenever job details are shown)
  will just parse the JS from disk again. It'll be better to only do this when
  the file on disk changed.
- [x] Remove the current index and replace it with a redirect to `/jobs`
- [ ] API: only allow valid job & task status changes. Currently any change is allowed by the API.
- [ ] Frontend authentication
- [ ] User/Job Submission API authentication
- [ ] Auto-removal of old Workers
- [ ] Ensure "task state machine" can run in a single database transaction.
- [ ] Refactor `internal/manager/task_logs` so that it uses `internal/manager/local_storage`.

## Worker

- [x] OpenAPI client that actually runs as a process
- [x] Process lifecycle management, clean shutdown
- [x] Task executor
- [x] Command runner
- [x] Log uploader
- [x] Implement `frames-to-video` command.
- [x] Implement `move-directory` command.
- [x] Refactor CLI-running code by analyzing the current Blender and FFmpeg commands.
- [x] Re-queueing of task at shutdown of the main context (i.e. when pressing Ctrl+C). Since the context is closed, everything shuts down, but then it's too late to cleanly communicate to the Manager that the current task should be requeued.
- [ ] CLI option to override configured manager URL.
- [ ] Filesystem readability/writability tests
  - [ ] Add check per task-type
  - [ ] Allow workers to request blocklisting for certain job, with "reason" string.
- [ ] POSIX signal handling (sleep/wake up)
- [ ] Remove `doneChan` field from `Worker` struct, and the handling of it. Its intended use overlaps with the global context.

## Both

- [x] Worker registration & auto-reregistration
- [x] Worker API authentication
- [x] Worker sign on/off
- [x] Task status change handling
- [x] UPnP/SSDP for automatic Manager finding
- [x] Worker status change management, including lazy status changes
- [ ] Worker test mode
- [ ] Worker registration secret (to protect the Manager from arbitrary Workers registering)
- [x] Support the may-i-keep-running protocol thingy

## Blender Add-on

- [x] Communication with the Flamenco Manager API
- [x] Job construction
- [x] BAT-packing
- [x] uploading to Shaman
- [x] Editable job names (for rendering multiple jobs from the same filename)
- [ ] User authentication
- [ ] Add `options={'HIDDEN'}` to the Flamenco RNA properties.
- [ ] Add eTag-like checksum to the job types. The aim is to prevent artists from sending in a new job with an old job type. This can happen when the job type is changed on the Manager, but in Blender the "refresh" button wasn't pressed yet. This should only cover the job settings, not the rest of the compiler script.
- [ ] Frame Range: allow automatically syncing with scene frame start/end.
- [ ] Don't take output settings from the current scene, but have separate settings for Flamenco. To be fleshed out still.

## Nice to have / new features

- [ ] Notification system to push "job done" messages to. Ideally would be in a form/shape that allows sending a message to Rocket.Chat, Matrix, Telegram, Discord, email, webbrowser, push URL-encoded/JSON/XML to some URL, stuff like that. Idea by Dan McLaughlin.
- [ ] Notification client inside Blender itself, so that you get a message when your job is done.
- [ ] Separate the OpenAPI definition of Shaman from the rest of Flamenco Manager. That way a part of BAT can also use the code generator. It also is the first step towards running Shaman as a standalone service.
- [ ] Web frontend for configuration management
- [ ] Configuration provisioning support
- [ ] Sending data upstream to Flamenco Overwatch
- [ ] Video preview of jobs, with black frames for unrendered frames.
- [ ] "Manual approval" task between low-quality/preview renders and rendering the final quality
- [ ] Grid view of jobs / tasks, showing the latest-rendered image/video.
- [ ] Split up OpenAPI-generated server code into an interface per tag. This'll help splitting up the rather large `Flamenco` struct into smaller pieces.
- [ ] Job "required specs", to match with certain workers.
