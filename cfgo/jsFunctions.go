//+build js
package cfgo

import "syscall/js"

type funcResult struct {
	isError bool
	out js.Value
}

func asyncCall(fnName string, args ...interface{}) chan funcResult {
	scope := js.Global().Get("_golangScope")

	resultChan := make(chan funcResult)
	cb := js.ValueOf(js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		isErr := args[0].Int() == 1
		functionOutput := args[1]
		resultChan <- funcResult{isError: isErr, out: functionOutput}
		return 0
	}))

	argsNew := make([]interface{}, 0)
	argsNew = append(argsNew, cb)
	argsNew = append(argsNew, args...)
	scope.Get(fnName).Invoke(argsNew...)
	return resultChan
}