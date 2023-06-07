package main

import (
	"log"
	"net/http"
	"net/url"

	"gopkg.in/elazarl/goproxy.v1"
)

func main() {
	proxy := goproxy.NewProxyHttpServer()

	// Add a handler to modify the requests
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {

		log.Printf("Request received: %v", req.URL)

		// Hardcoding the cookie
		// Create the JSESSIONID cookie object
		cookie := &http.Cookie{
			Name:  "JSESSIONID",
			Value: "58C557BF13E7954202E8FF460DE08351",
			Path:  "/",
		}
		// Add the cookie to the request
		req.AddCookie(cookie)

		return req, nil
	})

	// Add a handler to modify the responses
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {

		// Print the hostnames of the requested URLs
		host := getHost(resp.Request.URL)
		log.Printf("Response received for %v: %v", host, resp.Status)

		// Check for cookies in the response and print them
		for _, cookie := range resp.Cookies() {
			log.Printf("Cookie received from %v: %v", host, cookie)
		}

		return resp
	})

	log.Fatal(http.ListenAndServe(":8080", proxy))
}

// Helper function to print the hostnames
func getHost(u *url.URL) string {
	if u != nil {
		return u.Host
	}
	return ""
}

// Run 2 terminals, "go run proxyserver.go" to run this proxy,
// and the other "curl google.de -L --proxy http://localhost:8080" to use this proxy to send http requests.

// "curl --user jogh275022 https://jira.elektrobit.com/rest/api/latest/issue/RAPTOR-1 -i | grep JSESSIONID"
// to get the session cookie to be hardcoded

// Test after hardcoding the cookie with "curl https://jira.elektrobit.com/rest/api/latest/issue/RAPTOR-1 -i --proxy http://localhost:8080"
// Status code 401, still needs authentication
