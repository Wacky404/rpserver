package main

import (
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/Wacky404/rpserver/internal/cmd"
	"github.com/joho/godotenv"
)

func main() {
	certFile := flag.String("cert", "/rpserver/certs/localhost.pem", "TLS certificate file")
	keyFile := flag.String("key", "/rpserver/certs/localhost-key.pem", "TLS key file")

	flag.Parse()
	godotenv.Load()

	if os.Getenv("JWT_SECRET") == "" {
		log.Println("a critical env var is not set!")
		os.Exit(1)
	}

	go func() {
		log.Println("HTTPS server is running on https://localhost:8443")
		err := cmd.ExecuteServer(":8443", *certFile, *keyFile)
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("HTTP server is running on http://localhost:8080")
	err := http.ListenAndServe(":8080", http.HandlerFunc(redirectToHTTPS))
	if err != nil {
		log.Fatal(err)
	}
}

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.URL.RequestURI()
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}
