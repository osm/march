package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"regexp"
)

// Config holds the configuration for march.
type Config struct {
	Port      string `json:"port"`
	Database  string `json:"database"`
	Archivers []Archiver
	Archives  []Archive
}

// Archive contains information about an archive.
type Archive struct {
	Name    string `json:"name"`
	Storage string `json:"storage"`
	Users   []struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
}

// Archiver contains the structure of an archiver script and the regexp that
// will be used to match the incoming URL.
type Archiver struct {
	Name   string `json:"name"`
	Script string `json:"script"`
	Regexp string `json:"regexp"`
	regexp *regexp.Regexp
}

// loadConfig loads the given configuration file into the app.
func (app *app) loadConfig(configFile string) error {
	// Load the contents of the file into memory.
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Errorf("error: unable to read config", err)
	}

	// Reset the archives and archivers in the app.
	app.archives = make(map[string]Archive)
	app.archivers = make([]Archiver, 0)

	// Create a new empty config object and unmarshal the contents of the
	// file into it.
	config := Config{}
	json.Unmarshal(bytes, &config)

	// Iterate over the archives, we'll perform a check to make sure that
	// we haven't duplicated archives.
	for _, a := range config.Archives {
		if _, ok := app.archives[a.Name]; ok {
			return fmt.Errorf("error: duplicate archive name:", a.Name)
		}
		app.archives[a.Name] = a
	}

	// Iterate over the archivers, for each archiver we'll compile a
	// regexp so that we can use it easily in a later stage. If the regexp
	// fails to compile we'll return an error.
	for _, a := range config.Archivers {
		a.regexp, err = regexp.Compile(a.Regexp)
		if err != nil {
			return fmt.Errorf("error: unable to compile regexp:", a.Regexp)
		}
		app.archivers = append(app.archivers, a)
	}

	// Store the simpler values.
	app.port = config.Port
	app.dbPath = config.Database

	// Everything was OK.
	return nil
}
