package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// md5sum takes the given file and calculates a md5sum for it.
func md5sum(filePath string) (string, error) {
	// Try to open the file.
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Initialize a new hash.
	hash := md5.New()

	// Copy the file into the hash.
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	// Return the sum as a hex string.
	return hex.EncodeToString(hash.Sum(nil)[:16]), nil
}
