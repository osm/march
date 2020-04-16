package main

import (
	"os"
	"os/exec"
	"path"
	"strings"
)

// getArchiveFromURL tries to extract an archive from the URL, if found we'll
// return the archive.
func (app *app) getArchiveFromURL(url string) *Archive {
	// Split the URL on slashes.
	parts := strings.Split(url[1:], "/")

	// Parts with a length less than one means no archive name in the URL.
	if len(parts) < 1 {
		return nil
	}

	// Give the parts a name, to improve code readability.
	name := parts[0]

	// Make sure that the archive exists before we proceed.
	archive, ok := app.archives[name]
	if !ok {
		return nil
	}

	return &archive
}

// getArchiveItemIDFromURL tries to extract an archive item ID from the given
// URL, if it exists we'll return a pointer to a file description that
// contains the requested file.
func (app *app) getArchiveItemFromURL(url string) *os.File {
	// Split the URL on slashes.
	parts := strings.Split(url[1:], "/")

	// The URL should contain two parts when splitted, the first part
	// should be the archive name and the second part is the id of the
	// archived item.
	if len(parts) != 2 {
		return nil
	}

	// Give the parts a name, to improve code readability.
	name := parts[0]
	id := parts[1]

	// Make sure that the archive exists before we proceed.
	if _, ok := app.archives[name]; !ok {
		return nil
	}

	// ... and make sure that the given ID is a valid UUID.
	if !isUUID(id) {
		return nil
	}

	// ... and verify that the file actually exists in the database.
	fileID := app.getFileIDByID(name, id)
	if fileID == "" {
		return nil
	}

	// And finally, make sure that the physical file actually exists.
	file, err := os.Open(path.Join(app.archives[name].Storage, fileID))
	if err != nil {
		return nil
	}

	return file
}

// archive downloads and stores the given URL in the archive.
func (app *app) archive(archive *Archive, url, id string) {
	// Iterate over all archivers, match the URL to the regexp and when a
	// match is successfull we'll execute the archiver script and end the
	// routine.
	for _, ar := range app.archivers {
		if ar.regexp.Match([]byte(url)) {
			// Construct a path of the file we want to archive.
			file := path.Join(archive.Storage, id)

			// Execute the script, all scripts should accept the
			// URL as first parameter and the output file as
			// second parameter.
			cmd := exec.Command(ar.Script, url, file)
			if err := cmd.Run(); err != nil {
				app.logger.Printf("failed to execute script: %s, %s, %s, raw error: %w", ar.Script, url, file, err)
				return
			}

			// Calculate a md5sum of the archived file.
			md5sum, err := md5sum(file)
			if err != nil {
				app.logger.Printf("md5sum failed for %s", file)
				return
			}

			// Fetch the file ID based on the md5sum.
			// If the fileID exists we already have the file
			// stored in our archive, so we don't need to store it
			// again. We do however want to store the id that was
			// returned to the requestor.
			// So we link the same file id to another id in the
			// database.
			fileID := app.getFileIDByMD5Sum(archive.Name, md5sum)
			if fileID == "" {
				fileID = id
			} else {
				os.Remove(file)
			}

			// Add the file to the archive.
			if err := app.addToArchive(id, fileID, archive.Name, url, md5sum); err != nil {
				app.logger.Printf("add to archive failed: %w", err)
				return
			}

			return
		}
	}
}
