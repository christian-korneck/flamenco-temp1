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
    ]
};


function compileJob(job) {
    const settings = job.settings;

    const echoTask = author.Task("echo", "misc");
    echoTask.addCommand(author.Command("echo", {message: settings.message}));

    const sleepTask = author.Task("sleep", "misc")
    sleepTask.addCommand(author.Command("sleep", {duration_in_seconds: settings.sleep_duration_seconds}))
    sleepTask.addDependency(echoTask); // Ensure sleeping happens after echo, and not at the same time.

    job.addTask(echoTask);
    job.addTask(sleepTask);
}
