package main

import (
 cf "flarego"
 "fmt"
 "strings"
)

func HomeDemo(req *cf.Request) cf.Response {
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

 return cf.BuildResponse().SetBody(out).Build()
}

func OriginDemo(req *cf.Request) cf.Response {
 const origin = "https://welcome.developers.workers.dev"
 const replaceStr = "Welcome to a serverless execution environment"

 out := cf.Fetch(origin, "GET", nil, nil)
 out = strings.Replace(out, replaceStr, "<p>Welcome <h3>to...</h3>" +
 "<h2>Golang!</h2></p> <br/> ", 1)
 return cf.BuildResponse().AddHeader("content-type", "text/html").SetBody(out).Build()
}

func main() {
 cf.Router.Add("/", HomeDemo)
 cf.Router.Add("/fetch-demo", OriginDemo)
 cf.Router.Run()
}