// TODO: Left off fixing this: 2025/08/03 15:49:13 error occured when trying to determine if migration needed: failed to check pending migrations: file does not exist
package main

import (
	"database/sql"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/Wacky404/rpserver/db"
	"github.com/Wacky404/rpserver/internal/cmd"
	"github.com/joho/godotenv"

	"github.com/golang-migrate/migrate"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	_ "github.com/lib/pq"
)

func main() {
	certFile := flag.String("cert", "/rpserver/certs/localhost.pem", "TLS certificate file")
	keyFile := flag.String("key", "/rpserver/certs/localhost-key.pem", "TLS key file")
	forceMigrate := flag.Bool("migrate", false, "Force run migrations")

	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Fatalf("Failed to load .env vars: %v", err)
	}

	if os.Getenv("JWT_SECRET") == "" {
		log.Println("a critical (JWT_SECRET) env var is not set!")
		os.Exit(1)
	} else if os.Getenv("DATABASE_URL") == "" {
		log.Println("a critical (DATABASE_URL) env var is not set!")
		os.Exit(1)
	} else if os.Getenv("DATABASE_PROVIDER") == "" {
		log.Println("a critical (DATABASE_PROVIDER) env var is not set!")
		os.Exit(1)
	}

	var (
		dbUrl      = string(os.Getenv("DATABASE_URL"))
		dbProvider = string(os.Getenv("DATABASE_PROVIDER"))
	)

	conn, err := sql.Open(dbProvider, dbUrl)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatalf("Failed to establish connection with DB: %v", err)
	}

	// creating migrator
	m, err := db.NewMigrator(conn, dbProvider, dbUrl)
	if err != nil {
		log.Fatalf("Failed to create new Migrator: %v", err)
	}
	defer m.Close()

	if *forceMigrate {
		log.Println("Force running migrations...")
		if err := m.Migrate.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			log.Fatalf("failed to migrate up: %v", err)
		}
	} else {
		if err := m.MigrateIfNeeded(); err != nil {
			log.Fatalf("error occured when trying to determine if migration needed: %v", err)
		}
	}

	go func() {
		log.Println("HTTPS server is running on https://localhost:8443")
		err := cmd.ExecuteServer(":8443", *certFile, *keyFile)
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("HTTP server is running on http://localhost:8080 to redirect traffic to HTTPS")
	err = http.ListenAndServe(":8080", http.HandlerFunc(redirectToHTTPS))
	if err != nil {
		log.Fatal(err)
	}
}

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.URL.RequestURI()
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}
