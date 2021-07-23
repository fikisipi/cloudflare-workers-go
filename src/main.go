package main

import (
 "fmt"
 "strings"
 "github.com/fikisipi/cloudflare-workers-go/cfgo"
)

func HomeDemo(req *cfgo.Request) {
 out := fmt.Sprintf(`
  ⚡ This is the home demo.
  ⚡ for the fetch demo, curl /fetch-demo
  ----------------------
  Pathname: %s
  Query params: %s
  Your continent: %s


  Headers:%s`,
  req.Pathname,
  printMap(req.QueryParams, "%s=%s\n"),
  req.Cf.Continent,
  printMap(req.Headers, "\n   - %s: %s"))

 req.Respond(out)
}

func FetchDemo(req *cfgo.Request) {
 const origin = "https://welcome.developers.workers.dev"
 const replaceStr = "Welcome to a serverless execution environment"

 out := cfgo.Fetch(origin, "GET", nil, nil)
 out = strings.Replace(out, replaceStr, "<p>Welcome <h3>to...</h3>" +
 "<h2>Golang!</h2></p> <br/> ", 1)
 req.Respond(out, cfgo.SetHeader("content-type", "text/html"))
}

var KV = cfgo.KV("PROBA")

var KeyValueDemo = func(req *cfgo.Request) {
 if v, has := req.QueryParams["value"]; has {
  KV.PutKey(req.QueryParams["key"], v)
 }
 currentVals := printMap(KV.ListKeyValues(""), " - proba[%s] = %s\n")
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

 req.Respond(resp, cfgo.SetHeader("content-type", "text/html"))
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