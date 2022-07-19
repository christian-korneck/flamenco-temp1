---
title: Worker Configuration
weight: 2
---

Flamenco Worker will read its configuration from `flamenco-worker.yaml` in the
*current working directory*.

This is an example of such a configuration file:

```yaml
manager_url: http://flamenco.local:8080/
task_types: blender, ffmpeg, file-management, misc
```

The task types are determined by the [job compiler scripts][scripts]. The ones
listed here are in use by the default scripts. These determine which kind of
tasks this Worker will get. See [task types][task-types] for more info.

[scripts]: {{< ref "usage/job-compiler-scripts/_index.md" >}}
[task-types]: {{< ref "usage/job-compiler-scripts/_index.md" >}}#task-types

## Worker Local Files

Apart from the above configuration file, which can be shared between Workers,
each Worker has a set of files that are specific to that Worker. These contain
the *worker credentials*, which are used to identify this worker to the Manager,
and a *database file* to queue task updates when the Manager is unreachable.

These files are stored in a platform-specific location:

| Platform | Default location |
| -- | -- |
| Linux | `$HOME/.local/share/flamenco` |
| Windows | `C:\Users\UserName\AppData\Local\Flamenco` |
| macOS | `$HOME/Library/Application Support/Flamenco` |

## Configuration from Environment Variables

Certain settings can be configured via environment variables.

- `FLAMENCO_HOME`: Directory for [Worker local files](#worker-local-files). If
  not given, the above defaults are used.
- `FLAMENCO_WORKER_NAME`: The name of the Worker. If not specified, the Worker
  will use the hostname.
