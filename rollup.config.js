//import resolve from '@rollup/plugin-node-resolve';
import {terser} from 'rollup-plugin-terser';
//import strip from '@rollup/plugin-strip';
//import copy from 'rollup-plugin-copy';
//import {minify} from 'html-minifier-terser';
//import babel from '@rollup/plugin-babel';
//import commonjs from '@rollup/plugin-commonjs';
//import { visualizer } from 'rollup-plugin-visualizer';
//import { SourceMapGenerator }  from 'source-map';
import acorn from 'acorn'
import walk from 'acorn-walk'
//import { SourceMapGenerator, SourceNode, SourceMapConsumer } from 'source-map';
import path from 'path'
import {exec} from 'child_process'
import * as recast from "recast";

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
                exec('C:\\Users\\Filip\\go\\tinygo\\bin\\tinygo build -o ../worker/module.wasm', {
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
            console.log('go build → worker/module.wasm')
        }
    }
}

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
            rewriteAst(),
            compileWasm(),
            terser({compress: {passes: 10}, ecma: 2015, format: {ecma: 2015, comments: false, indent_level: 0}}),
            //visualizer({sourcemap: true})
        ]
    }
}