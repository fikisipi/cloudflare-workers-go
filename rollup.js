var child_process = require('child_process');
var path = require('path')

try {
    require('rollup-plugin-terser')
}catch(err) {
    console.log('Dependencies from npm not installed.')
    console.log('Trying to run npm install...')
    child_process.spawnSync('npm', ['install'], {shell: true, stdio: ['inherit', 'inherit', 'inherit']})
}

var rollup = path.resolve(__dirname, 'node_modules', '.bin', 'rollup')
child_process.spawnSync(rollup, ['-c'], {shell: true, stdio: ['inherit', 'inherit', 'inherit']})