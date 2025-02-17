package rpserver

import (
	"log"
	"os"

	"github.com/Wacky404/rpserver/internal/cmd"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

func main() {
	if os.Getenv("JWT_SECRET") == "" {
		log.Println("a critical env var is not set!")
		os.Exit(1)
	}
	cmd.ExecuteServer()
}
