// +build js

package cfgo

// blah blagh
type KVNamespace struct {
	Name string
}

func KV(namespace string) KVNamespace {
	return KVNamespace{Name: namespace}
}

// Return the KV value based on the key
func (namespace KVNamespace) GetKey(key string) string {
	result := <- asyncCall("kvGet", namespace.Name, key)
	if result.isError || result.out.IsNull() { return "" }
	return result.out.String()
}

func (namespace KVNamespace) PutKey(key string, value string) {
	<- asyncCall("kvPut", namespace.Name, key, value)
}

func (namespace KVNamespace) PutKeyExpiring(key string, value string, seconds int) {
	opts := make(map[string]interface{})
	opts["expirationTtl"] = seconds
	<- asyncCall("kvPut", namespace.Name, key, value, opts)
}

func (namespace KVNamespace) ListKeyValues(prefix string) map[string]string{
	res := <- asyncCall("kvListValues", namespace.Name, prefix)
	sMap := make(map[string]string)
	if res.isError { return sMap }
	return getJsMap(res.out)
}
