package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/Wacky404/rpserver/internal/cmd"
	"github.com/joho/godotenv"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/cockroachdb"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"

	_ "github.com/golang-migrate/migrate/v4/database"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func main() {
	certFile := flag.String("cert", "/rpserver/certs/localhost.pem", "TLS certificate file")
	keyFile := flag.String("key", "/rpserver/certs/localhost-key.pem", "TLS key file")

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

	db, err := sql.Open(dbProvider, dbUrl)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// this ish not working for some reason; going through it
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to establish connection with DB: %v", err)
	}

	var driver database.Driver

	switch dbProvider {
	case "postgres":
		driver, err = postgres.WithInstance(db, &postgres.Config{})
	case "cockroachdb":
		driver, err = cockroachdb.WithInstance(db, &cockroachdb.Config{})
	case "sqlite3":
		driver, err = sqlite3.WithInstance(db, &sqlite3.Config{})
	default:
		log.Fatalf("Unsupported DB provided: %s", dbProvider)
	}

	if err != nil {
		log.Fatalf("Failed to create DB driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		// file:///absolute/path | file://relative/path
		"file:///migrations",
		dbProvider,
		driver,
	)
	if err != nil {
		log.Fatalf("Migration init failed: %v", err)
	}

	if err := m.Up(); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	}

	go func() {
		log.Println("HTTPS server is running on https://localhost:8443")
		err := cmd.ExecuteServer(":8443", *certFile, *keyFile)
		if err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("HTTP server is running on http://localhost:8080")
	err = http.ListenAndServe(":8080", http.HandlerFunc(redirectToHTTPS))
	if err != nil {
		log.Fatal(err)
	}
}

func redirectToHTTPS(w http.ResponseWriter, r *http.Request) {
	target := "https://" + r.Host + r.URL.RequestURI()
	http.Redirect(w, r, target, http.StatusMovedPermanently)
}
