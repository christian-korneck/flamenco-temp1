---
title: Job Compiler Scripts
weight: 10
---

Flamenco uses *job compiler scripts* to convert a job definition (like "*render
this Blend file*") into concrete tasks that can be executed by workers.

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
compiler scripts](built-in-scripts) as examples.

[built-in-scripts]: https://developer.blender.org/diffusion/F/browse/main/internal/manager/job_compilers/scripts/

## Task Types

In the above example code, you can see calls like `author.Task("echo", "misc")`.
The first parameter is the task name, which will be shown in the web interface
and can help to identify each task. The second parameter is the *task type*.

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

In the end, these are just strings. When you create your own job compiler
scripts, you can follow these types or make them up yourself. Don't forget to
configure your workers to run them!

[worker-config]: {{< ref "usage/worker-configuration" >}}
