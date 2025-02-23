package cmd

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/Wacky404/rpserver/internal/auth"
)

func ExecuteServer(portNum string) error {
    // don't know if want to keep handlers in here
	http.HandleFunc("/proxy", handleProxy)

	log.Printf("Reverse Proxy running on :%v", portNum)
	err := http.ListenAndServe(":"+portNum, nil)

	return err
}

func handleProxy(w http.ResponseWriter, r *http.Request) {
	if !auth.Verifyrequest(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	backendURL, err := getBackendURL(r)
	if err != nil {
		http.Error(w, "Backend URL not provided", http.StatusBadRequest)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(backendURL)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	r = r.WithContext(ctx) // attaching the new ctx to request

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = backendURL.Scheme
		req.URL.Host = backendURL.Host
		req.URL.Path = backendURL.Path
		req.Host = backendURL.Host
	}

	done := make(chan struct{})
	go func() {
		proxy.ServeHTTP(w, r) // Forward the request
		close(done)
	}()

	select {
	case <-ctx.Done(): // if context timout occurs
		http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		log.Println("Request to", r.URL.Path, "timed out...balls")
	case <-done: // if request completes successfully
	}
}

func getBackendURL(r *http.Request) (*url.URL, error) {
	backend := r.Header.Get("X-Backend-URL") // extraction of backend from request

	if backend == "" {
		return nil, http.ErrNoLocation // error if no backend is in request
	}

	return url.Parse(backend) // parse and return the backend url
}
