//+build js
package structs

import "syscall/js"

type StringBody struct { Value string }
func (body StringBody) Get() js.Value {
	return js.ValueOf(body.Value)
}

type FormBody struct { Value map[string]string }
func (body FormBody) Get() js.Value {
	return CreateJsMap(body.Value)
}
