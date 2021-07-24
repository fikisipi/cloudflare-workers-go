package cfgo

func ExampleRequest_Respond() {
	var m *Request
	m.Respond("My body", SetStatus(403), SetHeader("x-a", "val"))
}
