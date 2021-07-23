# 👷 CloudFlare Workers in Go

`cfgo` (<a href="https://workers.cloudflare.com/">CloudFlare Workers</a> in Go) uses WebAssembly to bring
Go projects to the Workers.

To set up a project, install [CloudFlare Wrangler](https://github.com/cloudflare/wrangler) and run:

```
wrangler generate yourapp https://github.com/fikisipi/cloudflare-workers-go
```
### 🚴 Example and deployment
A demo with request handling is available in  `src/main.go`.

Run it using `wrangler dev`. To deploy
live, use `wrangler publish`.

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