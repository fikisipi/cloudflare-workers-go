import wasmGo from './go-1.16/wasm_exec.js'

const supportedNamespaces = () => {
    return Reflect.ownKeys(global).filter(x => {
        if (x !== 'origin' && typeof global[x] == 'object') {
            return global[x].constructor.name === 'KvNamespace'
        }
        return false;
    });
}

const serializeReq = (request) => {
    const url = request.url;
    const parsed = new URL(url)
    const reqObj = {
        Body: request.body,
        Cf: request.cf,
        Headers: Object.fromEntries(request.headers.entries()),
        Method: request.method,
        URL: request.url,
        Hostname: parsed.hostname,
        Pathname: parsed.pathname,
        QueryParams: Object.fromEntries(parsed.searchParams.entries())
    }
    return reqObj;
}

const dataQueue = [];

global['_doHandShake'] = () => {
    const { requestBlob, responseFunction } = dataQueue.pop();
    return { requestBlob, responseFunction };
}

const putHandshake = (requestBlob, responseFunction) => {
    const handshake = {requestBlob, responseFunction};
    dataQueue.push(handshake);
}

const WASM_FETCH = '_cfFetch';
global[WASM_FETCH] = (url, method, headers, body, cb) => {
    new Promise(async (resolve, reject) => {
        const initObj = {method};
        if (body != null) {
            if (typeof body == 'object') {
                const formData = new FormData();
                [...body.entries()].map(x => formData.append(x[0], x[1]))
                initObj.body = formData
            } else {
                initObj.body = body
            }
        }
        if (headers != null && Object.keys(headers).length > 0) {
            initObj.headers = new Headers(headers)
        }
        let resp = await fetch(url, initObj);
        let text = await resp.text();
        resolve(text)
    }).then(x => cb(x))
    return true
}

const golangFunctions = {
    async kvGet(namespace, key) {
        const value = await global[namespace].get(key)
        return value
    },
    async kvGetMeta(namespace, key) {
        return await global[namespace].getWithMetadata(key)
    },
    async kvPut(namespace, key, value, options) {
        return await global[namespace].put(key, value, options)
    },
    async kvDelete(namespace, key) {
        return await global[namespace].delete(key)
    },
    async kvList(namespace, prefix) { // cursors?
        return (await global[namespace].list({prefix: prefix})).keys
    },
    async kvListValues(namespace, prefix) {
        var out = {};
        const keyNames = (await golangFunctions.kvList(namespace, prefix)).map(x => x.name)
        for(const x of keyNames) {
            out[x] = await golangFunctions.kvGet(namespace, x)
        }
        return out;
    }
}

const GOLANG_SCOPE = '_golangScope';
global[GOLANG_SCOPE] = Object.fromEntries([...Object.entries(golangFunctions)].map((kv) => {
    const name = kv[0];
    const asyncFun = kv[1];

    return [name, (...args) => {
        const [callback, ...rest] = args;
        const IS_ERROR = 1;
        const IS_OK = 0;
        asyncFun(...rest).then(result => callback(IS_OK, result)).catch(result => {
            // We want a switch for this? Do JS errors go to logs?
            console.error(result)

            callback(IS_ERROR, result);
        })
    }];
}))

addEventListener('fetch', ev => {
    const requestBlob = serializeReq(ev.request)
    const programOutput = new Promise((resolve, reject) => {
        putHandshake(requestBlob, (response) => {
            resolve(response)
        })
    });

    const go = new wasmGo.Go()
   // go.argv.push(requestBlob)

    let instance = new WebAssembly.Instance(WASM_MODULE, go.importObject)

    ev.respondWith(async function () {
        let invocation = go.run(instance)
        // Race between program output & Worker limit timeout
        const winner = await Promise.race([programOutput, invocation])

        if (typeof winner === 'object') {
            let responseObj = winner
            const response = new Response(responseObj.Body,
                {
                    status: responseObj.StatusCode, headers: new Headers(responseObj.Headers)
                })
            return response
        } else {
            return new Response("Failed getting WASM response", {status: 500})
        }
    }())
})