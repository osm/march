package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// notFound returns a not found http error.
func (app *app) notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Found", http.StatusNotFound)
}

// noContent returns a no content http error.
func (app *app) noContent(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "No Content", http.StatusNoContent)
}

// unauthorized returns an unauthorized http error.
func (app *app) unauthorized(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

// badRequest returns a bad request http error.
func (app *app) badRequest(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Bad Request", http.StatusBadRequest)
}

// router handles the HTTP requests coming into the app.
func (app *app) router(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		app.get(w, r)
	} else if r.Method == "POST" {
		app.post(w, r)
	}
}

// get handles incoming GET requests, if the requested item exists in the
// archive we'll return it, otherwise 404.
func (app *app) get(w http.ResponseWriter, r *http.Request) {
	// Fetch the archived item.
	file, fileExists := app.getArchiveItemFromURL(r.URL.Path)
	if file == nil && !fileExists {
		app.notFound(w, r)
		return
	}
	if file == nil && fileExists {
		app.noContent(w, r)
		return
	}

	// Return it!
	http.ServeContent(w, r, "", time.Time{}, file)
}

// isAuthorized takes the given request and archive and checks whether or not
// the basic auth credentials are valid.
func (app *app) isAuthorized(r *http.Request, archive *Archive) bool {
	// Split the authorization header, the first part should just be the
	// header name, the second part contains the base64 encoded
	// credentials.
	parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
	if len(parts) != 2 {
		return false
	}

	// Decode the credentials, if it fails we'll just mark that as an
	// unauthorized request.
	bytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return false
	}

	// The credentials are separated by a :, that means that the first
	// part contains the username and second part is the password. This
	// should never fail, but just to make sure we'll check for it.
	creds := strings.SplitN(string(bytes), ":", 2)
	if len(creds) != 2 {
		return false
	}
	username := creds[0]
	password := creds[1]

	// Iterate over all the users in the archive and compare the
	// credentials, if they match we'll return true.
	for _, u := range archive.Users {
		if u.Username == username && u.Password == password {
			return true
		}
	}

	// Nope, the request was not authorized.
	return false
}

// post handles incoming POST requests, all requests are protected by basic
// auth, so first of all we'll make sure that the credentials are valid before
// anything is archived.
func (app *app) post(w http.ResponseWriter, r *http.Request) {
	// We'll start off by setting the basic auth header.
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

	// We'll hide the information about whether or not the archive exists
	// here, if we can't find it we'll just return it as an unauthorized
	// error.
	archive := app.getArchiveFromURL(r.URL.Path)
	if archive == nil {
		app.unauthorized(w, r)
		return
	}

	// Authorize the request.
	if !app.isAuthorized(r, archive) {
		app.unauthorized(w, r)
		return
	}

	// The request should have a url in the body, if it doesn't we'll
	// return a bad request.
	if err := r.ParseForm(); err != nil {
		app.badRequest(w, r)
		return
	}
	url := r.FormValue("url")
	if url == "" {
		app.badRequest(w, r)
		return
	}

	// Everything seems to be in order, let's generate a new id which will
	// be returned to the caller and launch the archive method in a new
	// goroutine.
	id := newUUID()
	go app.archive(archive, url, id)

	// Write the id back to the requestor.
	fmt.Fprintf(w, id+"\n")
}
