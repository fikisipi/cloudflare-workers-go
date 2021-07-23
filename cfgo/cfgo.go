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
func ResponseNew(body string) Response {
	reply := new(structs.RawResponse)

	reply.Body = body
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
		return 1
	})
	js.Global().Call("_cfFetch", url, method, headersJs, bodyJs, cb)
	return (<- out)
}

type RouteHandler struct {
	callbacks map[string]Callback
}

type Callback func(Request) Response

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
type JsonValue fastjson.Value
func (jVal *JsonValue) String() string {
	newval := (*fastjson.Value)(jVal)
	return string(newval.GetStringBytes())
}

func GetKey(namespace string, key string) string {
	result := <- asyncCall("kvGet", namespace, key)
	if result.isError || result.out.IsNull() { return "" }
	return result.out.String()
}

func PutKey(namespace string, key string, value string) {
	<- asyncCall("kvPut", namespace, key, value)
}

func PutKeyExpiring(namespace string, key string, value string, seconds int) {
	opts := make(map[string]interface{})
	opts["expirationTtl"] = seconds
	<- asyncCall("kvPut", namespace, key, value, opts)
}

func ListKeyValues(namespace string, prefix string) map[string]string{
	res := <- asyncCall("kvListValues", namespace, prefix)
	sMap := make(map[string]string)
	if res.isError { return sMap }
	return structs.GetJsMap(res.out)
}

// Dispatches the current Request to the first matching route
// added by Add()
func (h *RouteHandler) Run() {
	handshakeData := js.Global().Call("_doHandShake")
	reqBlob := handshakeData.Get("requestBlob")
	responseCallback := handshakeData.Get("responseFunction")

	var request Request
	request.Hostname = reqBlob.Get("Hostname").String()
	request.Body = reqBlob.Get("Body").String()
	request.URL = reqBlob.Get("URL").String()
	request.Method = reqBlob.Get("Method").String()
	request.Pathname = reqBlob.Get("Pathname").String()

	request.Headers = structs.GetJsMap(reqBlob.Get("Headers"))
	request.QueryParams = structs.GetJsMap(reqBlob.Get("QueryParams"))

	var response = ResponseNew("Route not found: " + request.Pathname).SetStatus(404)

	for pathStr, pathHandler := range h.callbacks {
		if matched, _ := path.Match(pathStr, request.Pathname); matched {
			response = pathHandler(request)
		}
	}

	rawResp := response.(*structs.RawResponse)
	responseObj := make(map[string]interface{})
	responseObj["StatusCode"] = rawResp.StatusCode
	responseObj["Body"] = rawResp.Body
	responseObj["Headers"] = structs.CreateJsMap(rawResp.Headers)

	responseCallback.Invoke(responseObj)
}

var Router = RouteHandler{}