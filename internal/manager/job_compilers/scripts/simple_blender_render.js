// SPDX-License-Identifier: GPL-3.0-or-later

const JOB_TYPE = {
    label: "Simple Blender Render",
    settings: [
        // Settings for artists to determine:
        { key: "frames", type: "string", required: true, eval: "f'{C.scene.frame_start}-{C.scene.frame_end}'",
          description: "Frame range to render. Examples: '47', '1-30', '3, 5-10, 47-327'" },
        { key: "chunk_size", type: "int32", default: 1, description: "Number of frames to render in one Blender render task" },

        // render_output_root + add_path_components determine the value of render_output_path.
        { key: "render_output_root", type: "string", subtype: "dir_path", required: true,
          description: "Base directory of where render output is stored. Will have some job-specific parts appended to it"},
        { key: "add_path_components", type: "int32", required: true, default: 0, propargs: {min: 0, max: 32},
          description: "Number of path components of the current blend file to use in the render output path"},
        { key: "render_output_path", type: "string", subtype: "file_path", editable: false,
          eval: "str(Path(settings.render_output_root) / last_n_dir_parts(settings.add_path_components) / jobname / '{timestamp}' / '######')",
          description: "Final file path of where render output will be saved"},

        // Automatically evaluated settings:
        { key: "blender_cmd", type: "string", default: "{blender}", visible: false },
        { key: "blendfile", type: "string", required: true, description: "Path of the Blend file to render", visible: false },
        { key: "fps", type: "float", eval: "C.scene.render.fps / C.scene.render.fps_base", visible: false },
        {
            key: "images_or_video",
            type: "string",
            required: true,
            choices: ["images", "video"],
            visible: false,
            eval: "'video' if C.scene.render.image_settings.file_format in {'FFMPEG', 'AVI_RAW', 'AVI_JPEG'} else 'images'"
        },
        { key: "format", type: "string", required: true, eval: "C.scene.render.image_settings.file_format", visible: false },
        { key: "image_file_extension", type: "string", required: true, eval: "C.scene.render.file_extension", visible: false,
          description: "File extension used when rendering images; ignored when images_or_video='video'" },
        { key: "video_container_format", type: "string", required: true, eval: "C.scene.render.ffmpeg.format", visible: false,
          description: "Container format used when rendering video; ignored when images_or_video='images'" },
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

// Mapping from video container (scene.render.ffmpeg.format) to the file name
// extension typically used to store those videos.
const videoContainerToExtension = {
    "QUICKTIME": ".mov",
    "MPEG1": ".mpg",
    "MPEG2": ".dvd",
    "MPEG4": ".mp4",
    "OGG": ".ogv",
    "FLASH": ".flv",
};

function compileJob(job) {
    print("Blender Render job submitted");
    print("job: ", job);


    const renderOutput = renderOutputPath(job);
    job.settings.render_output_path = renderOutput;

    const finalDir = path.dirname(renderOutput);
    const renderDir = intermediatePath(job, finalDir);

    const settings = job.settings;
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

// Do field replacement on the render output path.
function renderOutputPath(job) {
    let path = job.settings.render_output_path;
    return path.replace(/{([^}]+)}/g, (match, group0) => {
        switch (group0) {
        case "timestamp":
            return formatTimestampLocal(job.created);
        default:
            return match;
        }
    });
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
                "--render-frame", chunk.replace("-", ".."), // Convert to Blender frame range notation.
            ]
        });
        task.addCommand(command);
        renderTasks.push(task);
    }
    return renderTasks;
}

function authorCreateVideoTask(settings, renderDir) {
    if (settings.images_or_video == "video") {
        print("Not authoring video task, render output is already a video");
    }
    if (ffmpegIncompatibleImageFormats.has(settings.format)) {
        print("Not authoring video task, FFmpeg-incompatible render output")
        return;
    }
    if (!settings.fps) {
        print("Not authoring video task, no FPS known:", settings);
        return;
    }

    const stem = path.stem(settings.blendfile).replace('.flamenco', '');
    const outfile = path.join(renderDir, `${stem}-${settings.frames}.mp4`);
    const outfileExt = guessOutputFileExtension(settings);

    const task = author.Task('create-video', 'ffmpeg');
    const command = author.Command("create-video", {
        input_files: path.join(renderDir, `*${outfileExt}`),
        output_file: outfile,
        fps: settings.fps,
    });
    task.addCommand(command);

    print(`Creating output video for ${settings.format}`);
    return task;
}

// Return file name extension, including period, like '.png' or '.mkv'.
function guessOutputFileExtension(settings) {
    switch (settings.images_or_video) {
    case "images":
        return settings.image_file_extension;
    case "video":
        const container = settings.video_container_format;
        if (container in videoContainerToExtension) {
            return videoContainerToExtension[container];
        }
        return "." + container.lower();
    default:
        throw `invalid setting images_or_video: "${settings.images_or_video}"`
    }
}
