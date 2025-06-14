package cmd

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/Wacky404/rpserver/internal/middleware"
	"github.com/Wacky404/rpserver/internal/users"
	"github.com/golang-jwt/jwt/v5"
)

var (
	proxyCache = make(map[string]*httputil.ReverseProxy)
	cacheMutex sync.RWMutex
)

func ExecuteServer(port string, cert string, key string) error {
	mux := http.NewServeMux()

	mux.Handle("/", middleware.Recover(http.HandlerFunc(serveLoginPage)))
	mux.Handle("/auth/login", middleware.Recover(http.HandlerFunc(handleLogin)))
	mux.Handle("/dashboard", middleware.Recover(middleware.Cookies(http.HandlerFunc(serveDashboard))))
	mux.Handle("/proxy", middleware.Recover(middleware.JWT(http.HandlerFunc(handleProxy))))
	mux.Handle("/status", middleware.Recover(http.HandlerFunc(handleStatus)))

	err := http.ListenAndServeTLS(port, cert, key, mux)
	return err
}

func serveDashboard(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/dash/dashboard.html"))
	tmpl.Execute(w, nil)
}

func serveLoginPage(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, nil)
}

/* This function doesn't have proper auth for login creds */
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// pull this out into auth function
	if username == "admin" && password == "password4321" {
		//token, err := auth.GenerateJWT(username, time.Hour)
		//if err != nil {
		//	log.Printf("JWT generation error: %v", err)
		//	http.Error(w, "Could not generate token:", http.StatusInternalServerError)
		//	return
		//}

		//w.Header().Set("Content-Type", "application/json")
		//fmt.Fprintf(w, `{"token": "%s"}`, token)

		//return
		newSID := users.SessionPrefix + users.GenID(16) // hash and store in sessions table
		cookie := &http.Cookie{
			Name:     middleware.AdmitCookies[0],
			Value:    newSID,
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Now().Add(time.Minute * 2),
		}
		http.SetCookie(w, cookie)
		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusOK)
	}
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintln(w, "200 OK")
}

func handleProxy(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("claims").(jwt.MapClaims)
	if !ok {
		fmt.Println("Is this failing...")
		http.Error(w, "Failed to get JWT claims", http.StatusInternalServerError)
		return
	}
	userID := claims["sub"]
	role := claims["role"]
	fmt.Printf("%v, %v", userID, role)

	backendURL, err := getBackendURL(r)
	if err != nil {
		fmt.Println("Is this failing...2")
		http.Error(w, "Backend URL not provided", http.StatusBadRequest)
		return
	}

	proxy := getOrCreateProxy(backendURL)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	r = r.WithContext(ctx)

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = backendURL.Scheme
		req.URL.Host = backendURL.Host
		req.URL.Path = backendURL.Path
		req.Host = backendURL.Host
	}

	done := make(chan struct{})
	go func() {
		proxy.ServeHTTP(w, r)
		close(done)
	}()

	select {
	case <-ctx.Done():
		http.Error(w, "Request timed out", http.StatusGatewayTimeout)
		log.Println("Request to", r.URL.Path, "timed out...balls")
	case <-done:
	}
}

func getBackendURL(r *http.Request) (*url.URL, error) {
	backend := r.Header.Get("X-Backend-URL")
	if backend == "" {
		return nil, http.ErrNoLocation
	}
	return url.Parse(backend)
}

func getOrCreateProxy(target *url.URL) *httputil.ReverseProxy {
	cacheMutex.RLock()
	proxy, exists := proxyCache[target.String()]
	cacheMutex.RUnlock()
	if exists {
		return proxy
	}

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	if proxy, exists = proxyCache[target.String()]; exists {
		return proxy
	}

	proxy = httputil.NewSingleHostReverseProxy(target)
	proxyCache[target.String()] = proxy

	return proxy
}
