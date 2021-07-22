package cfgo

import (
	"syscall/js"
	"path"
	"github.com/fikisipi/cloudflare-workers-go/cfgo/structs"
	"github.com/valyala/fastjson"
)

type Request struct {
	Body string
	Headers map[string]string
	QueryParams map[string]string
	URL string
	Hostname string
	Pathname string
	Method string
}

type Response = structs.Response

// Used to chain .SetStatus(), .AddHeader(), etc.
// together when creating a Response. Example:
//   return BuildResponse().SetStatus(200).SetBody("Hello").Build()
// The final .Build() is not mandatory, it's just for
// reducing ambiguity about the return type.
func BuildResponse() Response {
	reply := new(structs.RawResponse)

	reply.StatusCode = 200
	reply.Headers = make(map[string]string)
	return reply
}

type FetchBody interface {
	Get() js.Value
}

func BodyString(body string) FetchBody {
	return &structs.StringBody{body}
}

func BodyForm(body map[string]string) FetchBody {
	return &structs.FormBody{body}
}

// Fetches any URL using the fetch() Web API. Unlike browsers,
// CloudFlare workers miss some features like credentials.
//
// If you don't need headers or a request body, set them to nil:
//    Fetch(myUrl, "GET", nil, nil)
//
// To create a POST/PUT body, use cfgo.BodyForm() or cfgo.BodyString():
//    Fetch(myURL, "PUT", nil, cfgo.BodyForm(myDict))
func Fetch(url string, method string, headers map[string]string, requestBody FetchBody) string {
	if headers == nil {
		headers = make(map[string]string)
	}
	headersJs := structs.CreateJsMap(headers)

	bodyJs := js.Null()
	if requestBody != nil {
		bodyJs = requestBody.Get()
	}

	out := make(chan string)
	var cb js.Func
	cb = js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		out <- args[0].String()
		// cb.Release()
		return 1
	})
	js.Global().Call("_cfFetch", url, method, headersJs, bodyJs, cb)
	return (<- out)
}

type RouteHandler struct {
	callbacks map[string]Callback
}

type Callback func(*Request) Response

// Adds a route, specified by a path and a callback.
//   Router.Add("/yourPath/*", myFunc)
// where myFunc is a Callback accepting a Request. The path
// argument supports wildcards.
//
// Note: The route list  is ordered, and the
// first route that matches the request is used.
func (h *RouteHandler) Add(routePath string, routeCallback Callback) {
	if(h.callbacks == nil) {
		h.callbacks = make(map[string]Callback)
	}
	h.callbacks[routePath] = routeCallback
}

// Dispatches the current Request to the first matching route
// added by Add()
func (h *RouteHandler) Run() {
	handshakeData := js.Global().Call("_doHandShake")
	jsonRequest := handshakeData.Get("requestBlob")
	responseCallback := handshakeData.Get("responseFunction")


	blob, _ := fastjson.Parse(jsonRequest.String())
	var request = new(Request)
	request.Pathname = string(blob.GetStringBytes("Pathname"))
	request.Body = string(blob.GetStringBytes("Body"))
	request.URL = string(blob.GetStringBytes("URL"))
	request.Method = string(blob.GetStringBytes("Method"))
	request.Hostname = string(blob.GetStringBytes("Hostname"))
	hdr := blob.GetObject("Headers")

	request.Headers = make(map[string]string)
	request.QueryParams = make(map[string]string)


	hdr.Visit(func(key []byte, v *fastjson.Value) {
		request.Headers[string(key)] = string(v.GetStringBytes())
	})

	var response = BuildResponse()

	_ = path.Match

	/*
	for pathStr, pathHandler := range h.callbacks {
		if matched, _ := path.Match(pathStr, request.Pathname); matched {
			response = pathHandler(request)
		}
	}*/
	for pathStr, pathHandler := range h.callbacks {
		if pathStr == request.Pathname {
			response = pathHandler(request)
		}
	}

	rawResp := response.(*structs.RawResponse)
	arena := fastjson.Arena{}
	responseObj := arena.NewObject()
	responseObj.Set("StatusCode", arena.NewNumberInt(rawResp.StatusCode))
	responseObj.Set("Body", arena.NewString(rawResp.Body))
	headers := arena.NewObject()
	for k, v := range rawResp.Headers {
		headers.Set(k, arena.NewString(v))
	}
	responseObj.Set("Headers", headers)

	responseBytes := responseObj.MarshalTo(nil)
	responseStr := string(responseBytes)
	result := responseStr

	responseCallback.Invoke(result)
}

var Router = RouteHandler{}