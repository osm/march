package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/osm/flen"
)

// app holds the main structure of the application.
type app struct {
	archivers []Archiver
	archives  map[string]Archive
	dbPath    string
	db        *sql.DB
	logger    *log.Logger
	port      string
}

func main() {
	// Define and parse the command line flags.
	configFile := flag.String("config", "", "Config file")
	flen.SetEnvPrefix("MARCH")
	flen.Parse()

	// No config file means that we can't proceed, so let's exit with an
	// error message.
	if *configFile == "" {
		fmt.Println("error: you need to specify a -config file")
		os.Exit(1)
	}

	// Initialize an empty app object.
	app := app{
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}

	// Load the configuration into the app.
	err := app.loadConfig(*configFile)
	if err != nil {
		fmt.Println("error: unable to load config", err)
		os.Exit(1)
	}

	// Initialize the database.
	app.initDB()

	// Execute the HTTP server.
	http.HandleFunc("/", app.router)
	http.ListenAndServe(":"+app.port, nil)
}
