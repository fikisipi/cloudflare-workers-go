# 👷 CloudFlare Workers for Go <img src="https://github.com/fikisipi/cloudflare-workers-go/actions/workflows/main.yml/badge.svg" />

`cfgo` uses WebAssembly to run Go projects as CloudFlare Workers. It exposes the APIs,
patches the missing runtime functions and glues the compiler to the CloudFlare tools.

To set up a project, install [CloudFlare Wrangler](https://github.com/cloudflare/wrangler) and run:

```
wrangler generate yourapp https://github.com/fikisipi/cloudflare-workers-go
```
### 🚴 Example deployment
```
~ wrangler login
~ wrangler publish
go build → worker/module.wasm
created worker/main.js in 1.2s
✨  Build completed successfully!
✨  Successfully published your script to
 https://myproject.myaccount.workers.dev
```
This runs the demos available in  `src/main.go`.


### 🚧️ TODO
* [x] Event/Request handling API
* [x] fetch API
* [x] Handle wasm_exec from non-latest (<1.16) & tinygo 
* [x] KV for Workers API
   * TODO : add metadata and cursor pagination
* [ ] WebSocket API
* [ ] Support for streaming & bytes in fetch
* [ ] 💥 reducing worker size
   * code stripping? (already doing AST optimization in `wasm_exec`)
   * handwritten optimizations
   * stdlib optimizations? `net/http/roundtrip_js.go`, `reflect/*.go`