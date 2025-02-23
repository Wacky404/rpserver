package rpserver

import (
	"log"
	"os"

	"github.com/Wacky404/rpserver/internal/cmd"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	if os.Getenv("JWT_SECRET") == "" {
		log.Println("a critical env var is not set!")
		os.Exit(1)
	}
	log.Fatal(cmd.ExecuteServer("9090"))
}
