import wasmGo from './tinygo-0.19/wasm_exec.js'

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

const dataQueue = [];

global['_doHandShake'] = () => {
    const { requestBlob, responseFunction } = dataQueue.pop();
    return { requestBlob, responseFunction };
}

const putHandshake = (requestBlob, responseFunction) => {
    const handshake = {requestBlob, responseFunction};
    dataQueue.push(handshake);
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