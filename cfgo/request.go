package cfgo

import (
	"syscall/js"
	"github.com/fikisipi/cloudflare-workers-go/cfgo/structs"
)

type RequestCFInfo struct {
	Asn             string
	Colo            string
	Country         string
	HttpProtocol    string
	RequestPriority string
	TLSCipher       string
	TLSClientAuth   string
	TLSVersion      string
	City            string
	Continent       string
	Latitude        string
	Longitude       string
	PostalCode      string
	MetroCode       string
	Region          string
	RegionCode      string
	Timezone        string
}

func makeCfFromMap(data map[string]string) RequestCFInfo {
	// Uh, we could use Reflection or fastjson here...
	return RequestCFInfo{
		data["asn"], data["colo"], data["country"], data["httpProtocol"],
		data["requestPriority"], data["tlsCipher"], data["tlsClientAuth"],
		data["tlsVersion"], data["city"], data["continent"], data["latitude"],
		data["longitude"], data["postalCode"], data["metroCode"], data["region"],
		data["regionCode"], data["timezone"],
	}
}

type Request struct {
	Body           string
	Headers        map[string]string
	QueryParams    map[string]string
	URL            string
	Hostname       string
	Pathname       string
	Method         string
	Cf             RequestCFInfo
	_response      *responseStruct
	_calledRespond bool
}

func (r *Request) Respond(body string, options ...ResponseOption) {
	r._calledRespond = true
	r._response = buildResponse(body, options...)
}

func makeRequestFromJs(reqBlob js.Value) *Request {
	var request = new(Request)

	request.Hostname = reqBlob.Get("Hostname").String()
	request.Body = reqBlob.Get("Body").String()
	request.URL = reqBlob.Get("URL").String()
	request.Method = reqBlob.Get("Method").String()
	request.Pathname = reqBlob.Get("Pathname").String()

	request.Headers = structs.GetJsMap(reqBlob.Get("Headers"))
	request.QueryParams = structs.GetJsMap(reqBlob.Get("QueryParams"))

	cfMap := structs.GetJsMap(reqBlob.Get("Cf"))
	request.Cf = makeCfFromMap(cfMap)

	request._response = new(responseStruct)
	request._calledRespond = false

	return request
}
