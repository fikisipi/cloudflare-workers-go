// @ts-ignore
import wasmGo from './go-1.16/wasm_exec.js'
import wsdemo from './wsdemo';

declare global {
    class WebSocketPair {
        client: WebSocket;
        server: WebSocket;
    }
    interface ResponseInit {
        webSocket?: WebSocket
    }
    interface URLSearchParams {
        entries: () => Iterable<[string, string]>;
    }
    interface Headers {
        entries(): IterableIterator<[string, string]>;
    }
    const WASM_MODULE: WebAssembly.Module;
    interface CFEvent extends Event {
        request: CFRequest;
        respondWith: (callback: Promise<any>) => any;
    }
}

const supportedNamespaces: () => Array<any> = () => {
    return Reflect.ownKeys(global).filter(x => {
        if (x !== 'origin' && typeof global[x] == 'object') {
            return global[x].constructor.name === 'KvNamespace'
        }
        return false;
    });
}

const serializeReq = (request: CFRequest) => {
    const url = request.url;
    const parsed = new URL(url)
    return {
        Body: request.body,
        Cf: request.cf,
        Headers: Object.fromEntries(request.headers.entries()),
        Method: request.method,
        URL: request.url,
        Hostname: parsed.hostname,
        Pathname: parsed.pathname,
        QueryParams: Object.fromEntries(parsed.searchParams.entries())
    };
}

interface IncomingCF {
    asn: string,
    colo: string,
    country: string,
    httpProtocol: string,
    requestPriority: string,
    tlsCipher: string,
    tlsClientAuth: string,
    tlsVersion: string,
    city: string,
    continent: string,
    latitude: string,
    longitude: string,
    postalCode: string,
    metroCode: string,
    region: string,
    regionCode: string,
    timezone: string,
}

interface OutgoingCF {
    apps?: boolean,
    cacheEverything?: boolean,
    cacheKey?: string,
    cacheTtl?: number,
    cacheTtlByStatus?: {[key: string]: number},
    minify?: {javascript?: boolean, css?: boolean, html?: boolean},
    mirage?: boolean,
    polish?: string,
    resolveOverride?: string,
    scrapeShield?: string
}

interface CFRequest extends Request {
    constructor: (input: string|CFRequest, init?: RequestInit & {cf?: OutgoingCF}) => void;
    readonly cf: IncomingCF
}

interface Handshake {
    requestBlob: CFRequest,
    responseFunction: () => any;
}

const handshakeQueue: Array<Handshake> = [];

global['_doHandShake'] = () => {
    const { requestBlob, responseFunction } = handshakeQueue.pop();
    return { requestBlob, responseFunction };
}

const putHandshake = (requestBlob, responseFunction) => {
    const handshake = {requestBlob, responseFunction};
    handshakeQueue.push(handshake);
}

const WASM_FETCH = '_cfFetch';
global[WASM_FETCH] = (url, method, headers, body, cb) => {
    new Promise(async (resolve, reject) => {
        const initObj: RequestInit = {method};
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
    const name: string = kv[0];
    const asyncFun: Function = kv[1];

    return [name, (...args) => {
        const [callback, ...rest] = args;
        const IS_ERROR = 1;
        const IS_OK = 0;
        asyncFun(...rest).then(result => callback(IS_OK, result)).catch(result => {
            console.log("Got an error from Go scope:")
            // We want a switch for this? Do JS errors go to logs?
            console.error(result)

            callback(IS_ERROR, result);
        })
    }];
}))

const ws = async req => {
    /*
    console.log(req.headers)
    const h = req.headers.get("upgrade")
    if(h != "websocket") return new Response("no upgrade", {status: 400}) */

    const [client, server] = Object.values(new WebSocketPair())
    await handleSession(server)

    return new Response(null, {
        status: 101,
        webSocket: client
    })
}

async function handleSession(websocket) {
    websocket.accept()
    let count = 0
    websocket.addEventListener("message", async ({ data }) => {
        if (data === "CLICK") {
            count += 1
            websocket.send(JSON.stringify({ count, tz: new Date() }))
        } else {
            // An unknown message came into the server. Send back an error message
            websocket.send(JSON.stringify({ error: "Unknown message received", tz: new Date() }))
        }
    })

    websocket.addEventListener("close", async evt => {
        // Handle when a client closes the WebSocket connection
        console.log(evt)
    })
}

addEventListener('fetch', (ev: CFEvent) => {
    const requestBlob = serializeReq(ev.request)
    const programOutput = new Promise((resolve, reject) => {
        putHandshake(requestBlob, (response) => {
            resolve(response)
        })
    });

    const go = new wasmGo.Go()
    let instance = new WebAssembly.Instance(WASM_MODULE, go.importObject)

    ev.respondWith(async function () {
        if(requestBlob.Pathname == '/ws') {
            return await ws(ev.request);
        }
        if(requestBlob.Pathname == '/ws-demo') {
            return new Response(wsdemo, {headers: {'content-type': 'text/html'}});
        }
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