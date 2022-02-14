# Flamenco-NG Features

This file lists the features that ought to be implemented at some point in time.
It will likely be expanded over time while forgotten features of Flamenco v2 are
rediscovered.

## Manager

- [x] OpenAPI implementation that actually runs as a server process
- [x] Process lifecycle management, clean shutdown
- [ ] Task Scheduler
- [ ] Unify command definition (type/parameters in persistence layer, name/settings in API layer)
- [ ] Job status change handling
- [ ] Task timeout monitoring
- [ ] Worker Blacklisting & failed task requeueing
- [ ] Worker timeout monitoring
- [ ] Variable replacement
- [ ] Configuration file reading
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
- [ ] Task executor
- [ ] Command runner
- [ ] Log uploader
- [ ] POSIX signal handling (sleep/wake up)

## Both

- [x] Worker registration & auto-reregistration
- [x] Worker API authentication
- [x] Worker sign on/off
- [ ] Worker status change management, including lazy status changes
- [ ] Task status change handling
- [ ] Worker test mode
- [ ] Worker registration secret (to protect the Manager from arbitrary Workers registering)
- [ ] SSDP/UPnP for automatic Manager finding
## Blender Add-on

- [ ] User authentication
- [ ] Communication with the Flamenco Manager API
- [ ] Job construction
- [ ] BAT-packing & uploading to Shaman
