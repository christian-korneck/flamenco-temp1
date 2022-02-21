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

const JOB_TYPE = {
    label: "Simple Blender Render",
    settings: [
        { key: "blender_cmd", type: "string", default: "{blender}", visible: false },
        { key: "filepath", type: "string", subtype: "file_path", required: true },
        { key: "chunk_size", type: "int32", default: 1 },
        { key: "frames", type: "string", required: true },
        { key: "render_output", type: "string", subtype: "hashed_file_path", required: true },
        { key: "fps", type: "int32" },
        { key: "extract_audio", type: "bool", default: true },
        {
            key: "images_or_video",
            type: "string",
            required: true,
            choices: ["images", "video"],
            visible: false,
        },
        { key: "format", type: "string", required: true },
        { key: "output_file_extension", type: "string", required: true },
    ]
};


// Set of scene.render.image_settings.file_format values that produce
// files which FFmpeg is known not to handle as input.
const ffmpegIncompatibleImageFormats = new Set([
    "EXR",
    "MULTILAYER", // Old CLI-style format indicators
    "OPEN_EXR",
    "OPEN_EXR_MULTILAYER", // DNA values for these formats.
]);

function compileJob(job) {
    print("Blender Render job submitted");
    print("job: ", job);

    const settings = job.settings;

    // The render path contains a filename pattern, most likely '######' or
    // something similar. This has to be removed, so that we end up with
    // the directory that will contain the frames.
    const renderOutput = settings.render_output;
    const finalDir = path.dirname(renderOutput);
    const renderDir = intermediatePath(job, finalDir);

    const renderTasks = authorRenderTasks(settings, renderDir, renderOutput);
    const videoTask = authorCreateVideoTask(settings, renderDir);

    for (const rt of renderTasks) {
        job.addTask(rt);
    }
    if (videoTask) {
        // If there is a video task, all other tasks have to be done first.
        for (const rt of renderTasks) {
            videoTask.addDependency(rt);
        }
        job.addTask(videoTask);
    }
}

// Determine the intermediate render output path.
function intermediatePath(job, finalDir) {
    const basename = path.basename(finalDir);
    const name = `${basename}__intermediate-${formatTimestampLocal(job.created)}`;
    return path.join(path.dirname(finalDir), name);
}

function authorRenderTasks(settings, renderDir, renderOutput) {
    print("authorRenderTasks(", renderDir, renderOutput, ")");
    let renderTasks = [];
    let chunks = frameChunker(settings.frames, settings.chunk_size);
    for (let chunk of chunks) {
        const task = author.Task(`render-${chunk}`, "blender");
        const command = author.Command("blender-render", {
            exe: settings.blender_cmd,
            argsBefore: [],
            blendfile: settings.filepath,
            args: [
                "--render-output", path.join(renderDir, path.basename(renderOutput)),
                "--render-format", settings.format,
                "--render-frame", chunk,
            ]
        });
        task.addCommand(command);
        renderTasks.push(task);
    }
    return renderTasks;
}

function authorCreateVideoTask(settings, renderDir) {
    if (ffmpegIncompatibleImageFormats.has(settings.format)) {
        print("Not authoring video task, FFmpeg-incompatible render output")
        return;
    }
    if (!settings.fps || !settings.output_file_extension) {
        print("Not authoring video task, no FPS or output file extension setting:", settings)
        return;
    }

    const stem = path.stem(settings.filepath).replace('.flamenco', '');
    const outfile = path.join(renderDir, `${stem}-${settings.frames}.mp4`);

    const task = author.Task('create-video', 'ffmpeg');
    const command = author.Command("create-video", {
        input_files: path.join(renderDir, `*${settings.output_file_extension}`),
        output_file: outfile,
        fps: settings.fps,
    });
    task.addCommand(command);

    print(`Creating output video for ${settings.format}`);
    return task;
}