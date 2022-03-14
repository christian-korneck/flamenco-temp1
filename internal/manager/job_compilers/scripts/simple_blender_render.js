// SPDX-License-Identifier: GPL-3.0-or-later

const JOB_TYPE = {
    label: "Simple Blender Render",
    settings: [
        { key: "blender_cmd", type: "string", default: "{blender}", visible: false },
        { key: "blendfile", type: "string", required: true, description: "Path of the Blend file to render", visible: false },
        { key: "chunk_size", type: "int32", default: 1, description: "Number of frames to render in one Blender render task" },
        { key: "frames", type: "string", required: true, eval: "f'{C.scene.frame_start}-{C.scene.frame_end}'"},
        { key: "render_output", type: "string", subtype: "hashed_file_path", required: true },
        { key: "fps", type: "float", eval: "C.scene.render.fps / C.scene.render.fps_base" },
        { key: "extract_audio", type: "bool", default: true },
        {
            key: "images_or_video",
            type: "string",
            required: true,
            choices: ["images", "video"],
            visible: false,
            eval: "'video' if C.scene.render.image_settings.file_format in {'FFMPEG', 'AVI_RAW', 'AVI_JPEG'} else 'image'"
        },
        { key: "format", type: "string", required: true, eval: "C.scene.render.image_settings.file_format" },
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
            blendfile: settings.blendfile,
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

    const stem = path.stem(settings.blendfile).replace('.flamenco', '');
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