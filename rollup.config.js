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

function golang() {
    return {
        name: 'go-transform',
        transform(code, file) {
            if (file.indexOf('wasm.js') == -1) return;

            const acornParser = {
                parse(source) {
                    return acorn.parse(source, {ecmaVersion: 2020, sourceType: 'module', locations: true});
                }
            };

            const ast = recast.parse(code, {
                parser: acornParser,
                sourceFileName: file
            })

            walk.fullAncestor(ast.program, (node, state, parent) => {
                if (node.type == 'MemberExpression') {
                    const snip = code.substring(node.start, node.end)
                    if (snip === `WebAssembly.instantiate`) {
                        parent.reverse()
                        const if_i = parent.findIndex(x => x.type == 'IfStatement');
                        const ifStatement = parent[if_i]
                        const parentStatement = parent[if_i + 1]
                        parentStatement.body = parentStatement.body.filter(x => x != ifStatement)
                    }
                }
            })
            return recast.print(ast).code
            return
        }
    };
}

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
                exec('go build -o ../worker/module.wasm', {
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
            golang(),
            compileWasm(),
            terser({compress: {passes: 10}, ecma: 2015, format: {ecma: 2015, comments: false, indent_level: 0}}),
            //visualizer({sourcemap: true})
        ]
    }
}