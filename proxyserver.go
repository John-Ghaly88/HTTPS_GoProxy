package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"gopkg.in/elazarl/goproxy.v1"
)

func main() {
	proxy := goproxy.NewProxyHttpServer()

	// Add a handler to modify the HTTP requests
	proxy.OnRequest().DoFunc(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		log.Printf("The HTTP OnRequest func was executed")

		log.Printf("Request received: %v", req.URL)

		// Hardcoding the cookie
		// Create the JSESSIONID cookie object
		cookie := &http.Cookie{
			Name:  "JSESSIONID",
			Value: "0DE065DED373AD78064BED9FC744B416", //Update with the cookie value
			Path:  "/",
		}
		// Add the cookie to the request
		req.AddCookie(cookie)

		return req, nil
	})

	// Add a handler to modify the HTTP responses
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		log.Printf("The HTTP OnResponse func was executed")

		// Print the hostnames of the requested URLs
		host := getHost(resp.Request.URL)
		log.Printf("Response received for %v: %v", host, resp.Status)

		// Check for cookies in the response and print them
		for _, cookie := range resp.Cookies() {
			log.Printf("Cookie received from %v: %v", host, cookie)
		}

		return resp
	})

	// TLS configuration for accepting HTTPS connections
	certFile := "certificate.crt" // Update with your certificate file path
	keyFile := "private.key"      // Update with your private key file path

	// Manually load the SSL certificate and private key
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("Failed to load certificate and private key: %v", err)
	}

	// Create a custom TLS configuration with the loaded certificate
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Set up a custom dialer to use the TLS configuration
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}

	// Create a custom transport with the TLS configuration
	transport := &http.Transport{
		TLSClientConfig: tlsConfig,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, addr)
		},
	}

	// Set the CONNECT handler to handle HTTPS connections
	proxy.ConnectDial = func(network, addr string) (net.Conn, error) {
		return transport.DialContext(context.Background(), network, addr)
	}

	// Add a handler to modify the HTTPS requests
	proxy.OnRequest(goproxy.ReqHostMatches(regexp.MustCompile(".*"))).HandleConnectFunc(
		func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
			log.Printf("The HTTPS OnRequest func was executed")

			// Get the CONNECT request
			req := ctx.Req

			// Add the JSESSIONID cookie to the CONNECT request headers
			cookie := &http.Cookie{
				Name:  "JSESSIONID",
				Value: "0DE065DED373AD78064BED9FC744B416", //Update with the cookie value
				Path:  "/",
			}
			req.AddCookie(cookie)

			return goproxy.OkConnect, host
		},
	)

	// Add a handler to modify the HTTPS responses
	proxy.OnResponse().DoFunc(func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		log.Printf("The HTTPS OnResponse func was executed")

		// Print the hostnames of the requested URLs
		host := getHost(resp.Request.URL)
		log.Printf("Response received for CONNECT %v: %v", host, resp.Status)

		// Check for cookies in the response and print them
		for _, cookie := range resp.Cookies() {
			log.Printf("Cookie received from CONNECT %v: %v", host, cookie)
		}

		return resp
	})

	proxy.Verbose = true
	proxy.Tr = transport

	log.Fatal(http.ListenAndServe(":8080", proxy))
}

// Helper function to print the hostnames
func getHost(u *url.URL) string {
	if u != nil {
		return u.Host
	}
	return ""
}

// To use this proxy
// 1- Replace /path/to/certificate.crt and /path/to/privatekey.key in the code with the actual paths to your certificate and private key files.

// 2- Get the session cookie, then hardcode it and save

// 3- Run the proxy using this cmd "go run proxyserver.go"

// 4- Test with this cmd "curl https://google.de -L --proxy http://localhost:8080"

// 5- Use this cmd "curl https://jira.elektrobit.com/rest/api/latest/issue/RAPTOR-1 -i --proxy http://localhost:8080"
