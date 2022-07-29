---
title: Job Types
weight: 10
---

Flamenco uses JavaScript files called *job compiler scripts* or two things. They

- define a *job type*, with its own set of settings, and
- convert a job of this type (like "*render this Blend file*") into concrete
  tasks that can be executed by workers.

## Example

These scripts are written in JavaScript. Before describing how they work, here
is a simple example that just logs a message (one task) and sleeps for a bit
(another task):

```JavaScript
const JOB_TYPE = {
    label: "Log and Sleep",
    settings: [
        { key: "message", type: "string", required: true },
        { key: "sleep_duration_seconds", type: "int32", default: 1 },
    ]
};

function compileJob(job) {
    const settings = job.settings;

    const echoTask = author.Task("echo", "misc");
    echoTask.addCommand(author.Command("echo", {
        message: settings.message,
    }));
    job.addTask(echoTask);

    const sleepTask = author.Task("sleep", "misc")
    sleepTask.addCommand(author.Command("sleep", {
        duration_in_seconds: settings.sleep_duration_seconds,
    }))
    job.addTask(sleepTask);
}
```

Each job compiler has two parts to it:

- the `JOB_TYPE` that defines the settings and how it's shown in the Blender add-on, and
- the `compileJob()` function that takes those settings and creates the tasks.

TODO: write more documentation. For now, you can refer to [the built-in job
compiler scripts][built-in-scripts] as examples.

[built-in-scripts]: https://developer.blender.org/diffusion/F/browse/main/internal/manager/job_compilers/scripts/

## Task Types

Each Flamenco task has a *task type*. This is a broad indicator of the kind of
work this task encompasses. The task's type determines which worker is allowed
to execute the task.

The following task types are defined by the standard job compiler scripts:

- `blender`: any task that runs Blender.
- `ffmpeg`: any task that runs FFmpeg.
- `file-management`: moving or copying files, creating directories, moving old
  renders out of the way to make place for new ones, etc.
- `misc`: tasks mostly used for testing things, such as "log a debug message" or
  "do nothing for 5 seconds".

These can be used to determine which worker executes which task. For example,
let a puny machine do the file management and maybe video compression with
FFmpeg, while the Blender tasks are left for the big ones. By default the
workers can run all the above task types. See [worker
configuration][worker-config] for more info.

In the above example code, you can see calls like `author.Task("echo", "misc")`.
The first parameter is the task name, which will be shown in the web interface
and can help to identify each task. The second parameter is the *task type*.

In the end, these are just strings. When you create your own job compiler
scripts, you can follow these types or make them up yourself. Don't forget to
configure your workers to run them!

[worker-config]: {{< ref "usage/worker-configuration" >}}

## Job Settings

The `JOB_TYPE` object contains the *job settings*. These can be shown in
Blender's user interface for submitting files, but can also be automatically
calculated based on the current file, the [context][context], etc.

The exact specifications of these settings can be found via the "API" link in
the top-right corner of the Flamenco Manager web interface. Search for the
`AvailableJobSetting` schema. This will always be accurate for your particular
version of Flamenco.

TODO: write about these in more detail.

### Available Python names for Evaluation

Job settings can have an `eval` key, which is a Python expression that
determines the value of the setting. That Python expression has access to the
following names:

- `jobname`: The name of the current Flamenco job.
- `settings`: The settings of the current Flamenco job.
- `bpy`: Blender's Python interface ([docs][bpy])
- `C`: The current [context][context], short-hand for `bpy.context`.
- `Path`: The `Path` class from Python's [pathlib][pathlib] library for file &
  directory path manipulation. Note that this does *not* understand Blender's
  `//` prefix for blendfile-relative paths. Use `bpy.path.abspath()` to turn
  those into an absolute path if necessary.
- `last_n_dir_parts(n, Optional[file_path])`: a function that returns the last
  `n` directory parts of some file's path. For example,
  `last_n_dir_parts(2, '/complex/path/to/a/file.blend')` will return `to/a`, as
  those are the last `2` components of the directory. If `file_path` is
  ommitted, it uses the current blend file, i.e. `bpy.data.filepath`.


[bpy]: https://docs.blender.org/api/master/
[context]: https://docs.blender.org/api/master/bpy.context.html
[pathlib]: https://docs.python.org/3/library/pathlib.html
