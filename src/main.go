package main

import (
 "fmt"
 "strings"
 "github.com/fikisipi/cloudflare-workers-go/cfgo"
)

const KV_NAMESPACE = "wrangler1_demo"

func HomeDemo(req cfgo.Request) cfgo.Response {
 out := fmt.Sprintf(`
  ⚡ This is the home demo.
  ⚡ for the fetch demo, curl /fetch-demo
  ----------------------
  Pathname: %s
  Query params: %s

  Headers:%s`,
  req.Pathname,
  printMap(req.QueryParams, "%s=%s\n"),
  printMap(req.Headers, "\n   - %s: %s"))

 return cfgo.ResponseNew(out)
}

func FetchDemo(req cfgo.Request) cfgo.Response {
 const origin = "https://welcome.developers.workers.dev"
 const replaceStr = "Welcome to a serverless execution environment"

 out := cfgo.Fetch(origin, "GET", nil, nil)
 out = strings.Replace(out, replaceStr, "<p>Welcome <h3>to...</h3>" +
 "<h2>Golang!</h2></p> <br/> ", 1)
 return cfgo.ResponseNew(out).AddHeader("content-type", "text/html").Build()
}

var KeyValueDemo = func(request cfgo.Request) cfgo.Response {
 if v, has := request.QueryParams["value"]; has {
  cfgo.PutKey(KV_NAMESPACE, request.QueryParams["key"], v)
 }
 currentVals := printMap(cfgo.ListKeyValues(KV_NAMESPACE, ""), " - proba[%s] = %s\n")
 resp := `
  <pre>` + currentVals + `
  ---
  </pre>
  <form>
  <input name="key" placeholder="Key" /> 
  <input name="value" placeholder="Value" /> <br/>
  <input type="submit" value="Set" />
  </form>
  `

 return cfgo.ResponseNew(resp).AddHeader("content-type", "text/html")
}

func main() {
 cfgo.Router.Add("/", HomeDemo)
 cfgo.Router.Add("/fetch-demo", FetchDemo)
 cfgo.Router.Add("/kv", KeyValueDemo)

 cfgo.Router.Run()
}

func printMap(strMap map[string]string, format string) (output string) {
 output = ""
 for k, v := range strMap {
  output += fmt.Sprintf(format, k, v)
 }
 return
}