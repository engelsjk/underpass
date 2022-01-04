package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/engelsjk/underpass"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file")
	}
}

func main() {

	saveLog := flag.Bool("save_logs", false, "save logs to file")
	flag.Parse()

	if !*saveLog {
		underpass.Start()
	}

	f, err := os.OpenFile(
		os.Getenv("UNDERPASS_LOG"),
		os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("unable to open log file %s", os.Getenv("UNDERPASS_LOG"))
		os.Exit(1)
	}
	defer f.Close()

	u := &underpass.Underpass{Log: f}
	u.Start()
}
