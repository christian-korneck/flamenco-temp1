var path = require('path');

print('Blender Render job submitted');
print('job: ', job)
print('running on platform', process.platform);

// Determine the intermediate render output path.
function intermediate_path(render_path) {
    const basename = path.basename(render_path);
    const name = `${basename}__intermediate-${job.created}`;
    return path.join(path.dirname(render_path), name);
}


// The render path contains a filename pattern, most likely '######' or
// something similar. This has to be removed, so that we end up with
// the directory that will contain the frames.
const render_output = path.dirname(job.settings.render_output);
print('render output', render_output);
const final_dir = path.dirname(render_output);
print('final dir    ', final_dir);
const render_dir = intermediate_path(final_dir);
print('render dir   ', render_dir);

// for (var i = 0; i < 5; i++) {
//     create_task('task' + i, 'task' + i + ' description');
// }

print('done creating tasks');