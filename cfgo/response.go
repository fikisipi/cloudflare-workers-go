//+build js
package cfgo

import (
	"syscall/js"
	"github.com/fikisipi/cloudflare-workers-go/cfgo/structs"
)

type responseStruct struct {
	Body string
	StatusCode int
	Headers map[string]string
}

func (response *responseStruct) serialize() js.Value {
	obj := make(map[string]interface{})
	obj["StatusCode"] = response.StatusCode
	obj["Body"] = response.Body
	obj["Headers"] = structs.CreateJsMap(response.Headers)
	return js.ValueOf(obj)
}

type ResponseOption interface {
	Apply(responseStruct *responseStruct)
}

func SetStatus(code int) ResponseOption {
	return applyFn(func(response *responseStruct) {
		response.StatusCode = code
	})
}

func SetHeader(key string, value string) ResponseOption {
	return applyFn(func(response *responseStruct) {
		response.Headers[key] = value
	})
}

func buildResponse(body string, options ...ResponseOption) *responseStruct {
	response := new(responseStruct)
	response.Headers = make(map[string]string)
	response.StatusCode = 200
	response.Body = body
	for _, option := range options {
		option.Apply(response)
	}
	return response
}

// ... Glue code for using a single function as interface

type applyFn func(response *responseStruct)
func (apply applyFn) Apply(responseStruct *responseStruct) {
	apply(responseStruct)
}