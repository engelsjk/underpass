package main

import (
	"flag"
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

	u := &underpass.Underpass{}

	if err := u.InitDB(); err != nil {
		log.Fatal(err)
	}

	var f *os.File
	var err error

	if *saveLog {
		f, err = os.OpenFile(
			os.Getenv("UNDERPASS_LOG"),
			os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		defer f.Close()
	} else {
		f = os.Stderr
	}

	if err != nil {
		log.Fatal("error opening log file: %v", err)
	}

	u.InitRouter(f)

	if err := u.StartRouter(); err != nil {
		log.Fatal(err)
	}
}
