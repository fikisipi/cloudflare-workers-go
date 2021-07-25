// +build js

package cfgo

import "syscall/js"

func createJsMap(sMap map[string]string) js.Value {
	/* JS syscalls cannot work with map[string]string
	or any map[string]T as well.

	It must explicitly be map[string]interface{} because
	there's no generics and the recursive serializer must
	be able to recursively visit any value type.
	*/
	interfaceMap := make(map[string]interface{})
	for key, value := range sMap {
		interfaceMap[key] = value
	}
	return js.ValueOf(interfaceMap)
}

func getJsMap(value js.Value) map[string]string {
	m := make(map[string]string)
	entries := js.Global().Get("Object").Call("entries", value)
	for i := 0; i < entries.Length(); i++ {
		kv := entries.Index(i)
		key := kv.Index(0).String()
		value := kv.Index(1).String()
		m[key] = value
	}
	return m
}