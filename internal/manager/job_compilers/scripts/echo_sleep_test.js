// SPDX-License-Identifier: GPL-3.0-or-later


/* Example job JSON:

{
  "metadata": {
    "project": "Sprite Fright",
    "user.email": "sybren@blender.org",
    "user.name": "Sybren Stüvel"
  },
  "type": "echo-sleep-test",
  "name": "pošalji poruku i idi na spavanje",
  "priority": 50,
  "settings": {
    "message": "prespavati",
    "sleep_duration_seconds": 3
  }
}

*/

const JOB_TYPE = {
    label: "Echo Sleep Test",
    settings: [
        { key: "message", type: "string", required: true },
        { key: "sleep_duration_seconds", type: "int32", default: 1 },
        { key: "sleep_repeats", type: "int32", default: 1 },
    ]
};


function compileJob(job) {
    const settings = job.settings;

    const echoTask = author.Task("echo", "misc");
    echoTask.addCommand(author.Command("echo", {message: settings.message}));
    job.addTask(echoTask);

    for (let repeat=0; repeat < settings.sleep_repeats; repeat++) {
      const sleepTask = author.Task("sleep", "misc")
      sleepTask.addCommand(author.Command("sleep", {duration_in_seconds: settings.sleep_duration_seconds}))
      sleepTask.addDependency(echoTask); // Ensure sleeping happens after echo, and not at the same time.
      job.addTask(sleepTask);
    }
}
