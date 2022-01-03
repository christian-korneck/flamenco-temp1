var path = require('path');

print('Blender Render job submitted');
print('job: ', job)

const { created, settings } = job;

// Determine the intermediate render output path.
function intermediatePath(render_path) {
    const basename = path.basename(render_path);
    const name = `${basename}__intermediate-${created}`;
    return path.join(path.dirname(render_path), name);
}

function frameChunker(frames, callback) {
    callback('1-10');
    callback('11-20');
    callback('21-30');
}

// The render path contains a filename pattern, most likely '######' or
// something similar. This has to be removed, so that we end up with
// the directory that will contain the frames.
const renderOutput = path.dirname(settings.render_output);
const finalDir = path.dirname(renderOutput);
const renderDir = intermediatePath(finalDir);

let renderTasks = [];
frameChunker(settings.frames, function(chunk) {
    const task = author.Task(`render-${chunk}`);
    const command = author.Command('blender-render', {
        cmd: settings.blender_cmd,
        filepath: settings.filepath,
        format: settings.format,
        render_output: path.join(renderDir, path.basename(renderOutput)),
        frames: chunk,
    });
    task.addCommand(command);
    renderTasks.push(task);
});

print(`done creating ${renderTasks.length} tasks`);
for (const task of renderTasks) {
    print(task);
}