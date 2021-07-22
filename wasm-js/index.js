import wasmGo from './tinygo/wasm_exec.js'

const WASM_GET_CALLBACK = '_getCallback';
const WASM_FETCH = '_cfFetch';

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
    return JSON.stringify(reqObj)
}

const callbackQueue = [];

global[WASM_GET_CALLBACK] = () => {
    return callbackQueue.pop();
}

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

const putCallback = (fn) => {
    callbackQueue.push(fn);
}

addEventListener('fetch', ev => {
    const requestBlob = serializeReq(ev.request)
    const programOutput = new Promise((resolve, reject) => {
        putCallback((str) => {
            resolve(str)
        });
    });

    const go = new wasmGo.Go()
    go.argv.push(requestBlob)

    let instance = new WebAssembly.Instance(WASM_MODULE, go.importObject)

    ev.respondWith(async function () {
        let invocation = go.run(instance)
        // Race between program output & Worker limit timeout
        const winner = await Promise.race([programOutput, invocation])

        if (typeof winner == 'string') {
            let responseObj = JSON.parse(winner)
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