package cfgo

// blo
func ExampleRequest_Respond() {
	var m *Request
	// blah
	m.Respond("My body", SetStatus(403), SetHeader("x-a", "val"))
}

/*
<DOCMAP>
1. cfgo.go: CloudFlare
2. request.go: Requests
3. response.go: Responses
4. kv.go: KV (Key-Value) API
*/