package structs

import "syscall/js"

func CreateJsMap(sMap map[string]string) js.Value {
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
