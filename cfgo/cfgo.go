// +build js

package cfgo

import (
	"syscall/js"
	"path"
)

type FetchBody interface {
	Get() js.Value
}

type FormStruct struct {
	body map[string]string
}

func (s FormStruct) Get() js.Value {
	return createJsMap(s.body)
}

func BodyForm(body map[string]string) FetchBody {
	return &FormStruct{body}
}

// Fetches any URL using the fetch() Web API. Unlike browsers,
// CloudFlare workers miss some features like credentials.
//
// If you don't need headers or a request body, set them to nil:
//    Fetch(myUrl, "GET", nil, nil)
//
// To create a POST/PUT body, use cfgo.BodyForm() or cfgo.BodyString():
//    Fetch(myURL, "PUT", nil, cfgo.BodyForm(myDict))
func Fetch(url string, method string, headers map[string]string,
	requestBody FetchBody) string {
	if headers == nil {
		headers = make(map[string]string)
	}
	headersJs := createJsMap(headers)

	bodyJs := js.Null()
	if requestBody != nil {
		bodyJs = requestBody.Get()
	}

	out := make(chan string)
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		out <- args[0].String()
		return 1
	})
	js.Global().Call("_cfFetch", url, method, headersJs, bodyJs, cb)
	return (<- out)
}

type routeHandler struct {
	callbacks map[string]callback
}

type callback func(*Request)

// Adds a route, specified by a path and a callback.
//   Router.Add("/yourPath/*", myFunc)
// where myFunc is a Callback accepting a Request. The path
// argument supports wildcards.
//
// Note: The route list  is ordered, and the
// first route that matches the request is used.
func (h *routeHandler) Add(routePath string, routeCallback func(*Request)) {
	if(h.callbacks == nil) {
		h.callbacks = make(map[string]callback)
	}
	h.callbacks[routePath] = routeCallback
}

// Dispatches the current Request to the first matching route
// added by Add()
func (h *routeHandler) Run() {
	handshakeData := js.Global().Call("_doHandShake")
	reqBlob := handshakeData.Get("requestBlob")
	responseCallback := handshakeData.Get("responseFunction")

	request := makeRequestFromJs(reqBlob)

	var response = buildResponse("Route not found: " + request.Pathname, SetStatus(404))

	for pathStr, pathHandler := range h.callbacks {
		if matched, _ := path.Match(pathStr, request.Pathname); matched {
			pathHandler(request)

			if request._calledRespond {
				response = request._response
			} else {
				response = buildResponse("")
			}
		}
	}

	rawResp := response.serialize()

	responseCallback.Invoke(rawResp)
}

// Router Blah blah
var Router = routeHandler{}