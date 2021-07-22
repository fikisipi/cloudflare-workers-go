package main

import (
 "fmt"
 "strings"
 "github.com/fikisipi/cloudflare-workers-go/cfgo"
 "math/rand"
 "strconv"
)

func HomeDemo(req cfgo.Request) cfgo.Response {
 out := fmt.Sprintf(`
  ⚡ This is the home demo.
  ⚡ for the fetch demo, curl /fetch-demo
  ----------------------
  Pathname: %s
  Query params: %s

  Headers:%s`,
  req.Pathname,
  mapToStr(req.QueryParams, "%s=%s\n"),
  mapToStr(req.Headers, "\n   - %s: %s"))

 return cfgo.BuildResponse().SetBody(out).Build()
}

func FetchDemo(req cfgo.Request) cfgo.Response {
 const origin = "https://welcome.developers.workers.dev"
 const replaceStr = "Welcome to a serverless execution environment"

 out := cfgo.Fetch(origin, "GET", nil, nil)
 out = strings.Replace(out, replaceStr, "<p>Welcome <h3>to...</h3>" +
 "<h2>Golang!</h2></p> <br/> ", 1)
 return cfgo.BuildResponse().AddHeader("content-type", "text/html").SetBody(out).Build()
}

func main() {
 cfgo.Router.Add("/", HomeDemo)
 cfgo.Router.Add("/fetch-demo", FetchDemo)
 cfgo.Router.Add("/why", func(request cfgo.Request) cfgo.Response {
   return cfgo.BuildResponse().SetBody(strconv.Itoa(rand.Int()))
 })
 cfgo.Router.Run()
}

func mapToStr(strMap map[string]string, format string) (output string) {
 output = ""
 for k, v := range strMap {
  output += fmt.Sprintf(format, k, v)
 }
 return
}