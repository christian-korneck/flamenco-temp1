print('Blender Render job submitted');
print('job: ', job)

var pathlib = require('pathlib');
print('loaded module')
print('running pathlib.test()')
pathlib.test()

// The render path contains a filename pattern, most likely '######' or
// something similar. This has to be removed, so that we end up with
// the directory that will contain the frames.
const render_output = job.settings.render_output.replace(/[\\\/][^\\\/]*$/, '');
print('render output', render_output);
// final_dir = self.render_output.parent
// render_dir = intermediate_path(job, self.final_dir)



// for (var i = 0; i < 5; i++) {
//     create_task('task' + i, 'task' + i + ' description');
// }

print('done creating tasks');