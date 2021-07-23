import acorn from "acorn";
import * as recast from "recast";
import walk from "acorn-walk";

function rewriteAst(goVersion) {
    return {
        name: 'go-transform',
        transform(code, file) {
            if (file.indexOf('index.js') !== -1) {
                return code.replace("./go-1.16/wasm_exec.js", `./${goVersion}/wasm_exec.js`)
            }
            if (file.indexOf('wasm_') === -1) return;

            const acornParser = {
                parse(source) {
                    return acorn.parse(source, {ecmaVersion: 2020, sourceType: 'module', locations: true});
                }
            };

            const REF_IMPL = `syscall/js.finalizeRef not implemented`
            const SHOULD_PATCH = code.indexOf(REF_IMPL) !== -1;

            const ast = recast.parse(code, {
                parser: acornParser,
                sourceFileName: file
            })

            const bd = recast.types.builders;

            const reverse = (arr) => arr.slice().reverse();
            walk.fullAncestor(ast.program, (node, state, parents) => {
                if (node.type == 'MemberExpression') {
                    const snip = code.substring(node.start, node.end)
                    let name1, name2;
                    if (node.object.name) name1 = node.object.name;
                    if (node.property.name) name2 = node.property.name;
                    const expr = `${name1}.${name2}`;
                    const parent = reverse(parents);

                    // global.performance() needs to be patched for the
                    // vm CloudFlare supports
                    if (expr === `global.performance`) {
                        if (parent[2].type === 'IfStatement') {
                            parent[3].body = parent[3].body.filter(x => x !== parent[2])
                        }
                    }
                    // Remove the unneeded WASM instantiate code
                    if (expr === `WebAssembly.instantiate`) {
                        const if_i = parent.findIndex(x => x.type === 'IfStatement');
                        const ifStatement = parent[if_i]
                        const parentStatement = parent[if_i + 1]
                        parentStatement.body = parentStatement.body.filter(x => x !== ifStatement)
                    }

                    if(expr === 'console.error' && SHOULD_PATCH) {
                        if(parent[1].arguments && parent[1].arguments[0].value === REF_IMPL) {
                            //console.log(parent[5])
                            let patched = bd.identifier("_patchedRefFunction")
                            parent[5].value = bd.callExpression(patched, [bd.identifier("this"), bd.identifier("mem")])
                        }
                    }
                }
            })
            return recast.print(ast).code + `
			global.performance = { now: () => Date.now() }
			function _patchedRefFunction(obj, mem) {
                return (x) => {
                    const id = mem().getUint32(x, true);
                    obj._goRefCounts[id]--;
                    if (obj._goRefCounts[id] === 0) {
                        const v = obj._values[id];
                        obj._values[id] = null;
                        obj._ids.delete(v);
                        obj._idPool.push(id);
                    }
                }
            }
			export default { Go: global.Go }
			`;
        }
    };
}

export default rewriteAst;