package main

import (
 "fmt"
 "strings"
 "github.com/fikisipi/cloudflare-go/cfgo"
)

func HomeDemo(req *cfgo.Request) cfgo.Response {
 out := fmt.Sprintf(`
  ⚡ This is the home demo.
  ⚡ for the fetch demo, curl /fetch-demo
  ----------------------
  Pathname: %s
  Query params: %s`, req.Pathname, req.QueryParams)

 out += "\n\n  Headers:\n\n"
 for k, v := range req.Headers {
  out += fmt.Sprintf("  - %s: %s\n", k, v)
 }

 return cfgo.BuildResponse().SetBody(out).Build()
}

func OriginDemo(req *cfgo.Request) cfgo.Response {
 const origin = "https://welcome.developers.workers.dev"
 const replaceStr = "Welcome to a serverless execution environment"

 out := cfgo.Fetch(origin, "GET", nil, nil)
 out = strings.Replace(out, replaceStr, "<p>Welcome <h3>to...</h3>" +
 "<h2>Golang!</h2></p> <br/> ", 1)
 return cfgo.BuildResponse().AddHeader("content-type", "text/html").SetBody(out).Build()
}

func main() {
 cfgo.Router.Add("/", HomeDemo)
 cfgo.Router.Add("/fetch-demo", OriginDemo)
 cfgo.Router.Run()
}