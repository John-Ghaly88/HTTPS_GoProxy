package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"

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
			Value: "B79F26300BB44F22BEB0C5C1846B0028",
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

	// TLS configuration for accepting https connections
	certFile := "/path/to/certificate.crt" // Update with your certificate file path
	keyFile := "/path/to/privatekey.key"   // Update with your private key file path

	// 	Manually load the SSL certificate and private key
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

// Run 2 terminals, "go run proxyserver.go" to run this proxy,
// and the other "curl google.de -L --proxy http://localhost:8080" to use this proxy to send http requests.

// To use the proxy with a cookie:
// 1- Use this cmd to get the session cookie, then hardcode it
// "curl --user jogh275022 https://jira.elektrobit.com/rest/api/latest/issue/RAPTOR-1 -i | grep JSESSIONID"
// 2- Test after hardcoding the cookie with this cmd
// "curl http://jira.elektrobit.com/rest/api/latest/issue/RAPTOR-1 -i --proxy http://localhost:8080"

// To use the proxy with https:
// 1- Replace /path/to/certificate.crt and /path/to/privatekey.key in the code with the actual paths to your certificate and private key files.
// 2- Use this cmd "curl https://jira.elektrobit.com/rest/api/latest/issue/RAPTOR-1 -i --proxy https://localhost:8080 --proxy-insecure"
// The --proxy-insecure flag is necessary because your proxy server is using a self-signed certificate, and cURL will reject it by default,
// In a production setup, you should use a valid SSL certificate signed by a trusted certificate authority, and remove this flag.
