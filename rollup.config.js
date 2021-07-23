//import resolve from '@rollup/plugin-node-resolve';
import {terser} from 'rollup-plugin-terser';
//import strip from '@rollup/plugin-strip';
//import copy from 'rollup-plugin-copy';
//import {minify} from 'html-minifier-terser';
//import babel from '@rollup/plugin-babel';
//import commonjs from '@rollup/plugin-commonjs';
//import { visualizer } from 'rollup-plugin-visualizer';
//import { SourceMapGenerator }  from 'source-map';
//import { SourceMapGenerator, SourceNode, SourceMapConsumer } from 'source-map';
import path from 'path'
import {exec} from 'child_process'

//const exec = promisify(execOld)
import rewriteAst from "./wasm-ast";

let compileWasm = () => {
    return {
        name: 'compile-wasm',
        async buildEnd(err) {
            const arch = {
                ...process.env,
                GOARCH: 'wasm',
                GOOS: 'js'
            }
            let goResult = await new Promise((resolve, reject) => {
                console.log(`${GOEXEC} build → worker/module.wasm`)
                exec(`${GOEXEC} build -o ../worker/module.wasm`, {
                    cwd: process.cwd() + '/src',
                    env: arch
                }, (err, stdout, stderr) => {
                    stdout ? console.log(stdout) : null
                    stderr ? console.log(stderr) : null
                    if (err) {
                        console.error(err);
                        reject(err);
                    } else
                        resolve()
                })
            })
        }
    }
}

const GOEXEC = process.env.GO || 'go'
const GOVERSION = (GOEXEC === 'tinygo' ? 'tinygo-0.19' : 'go-1.16')

export default cmd => {
    return {
        input: './wasm-js/index.js',
        output: {
            file: 'worker/main.js',
            //format: 'es',
            format: 'iife',
            sourcemap: false,
            sourcemapFile: path.resolve('./custom-bundle.js.map')
        },
        plugins: [
            //resolve(),
            rewriteAst(GOVERSION),
            compileWasm(),
            terser({
                compress: {
                    passes: 10,
                    global_defs: {VERSION: 'tinygo'}
                }, ecma: 2015, format: {ecma: 2015, comments: false, indent_level: 0}
            }),
            //visualizer({sourcemap: true})
        ]
    }
}