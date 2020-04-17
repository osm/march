package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
	"github.com/osm/migrator"
	"github.com/osm/migrator/repository"
)

// getDBRepo returns a repository of the DB migrations.
func getDBRepo() repository.Source {
	return repository.FromMemory(map[int]string{
		1: `
			CREATE TABLE migration (
				version TEXT NOT NULL PRIMARY KEY
			);`,
		2: `
			CREATE TABLE archive_item (
				id VARCHAR(36) NOT NULL PRIMARY KEY,
				file_id VARCHAR(36) NOT NULL,
				archive TEXT NOT NULL,
				url TEXT NOT NULL,
				md5sum VARCHAR(32) NOT NULL,
				deleted_at DATETIME,
				created_at DATETIME NOT NULL
			);`,
		3: `
			CREATE INDEX archive_item_md5sum ON archive_item(md5sum);
		`,
		4: `
			CREATE TEMPORARY TABLE archive_item_backup (
				id VARCHAR(36) NOT NULL PRIMARY KEY,
				file_id VARCHAR(36) NOT NULL,
				archive TEXT NOT NULL,
				url TEXT NOT NULL,
				md5sum VARCHAR(32) NOT NULL,
				deleted_at DATETIME,
				created_at DATETIME NOT NULL
			);
			INSERT INTO archive_item_backup SELECT * FROM archive_item;
			DROP TABLE archive_item;
			CREATE TABLE archive_item (
				id VARCHAR(36) NOT NULL PRIMARY KEY,
				file_id VARCHAR(36),
				archive TEXT NOT NULL,
				url TEXT NOT NULL,
				md5sum VARCHAR(32) NOT NULL,
				deleted_at DATETIME,
				created_at DATETIME NOT NULL
			);
			INSERT INTO archive_item SELECT * FROM archive_item_backup;
			DROP TABLE archive_item_backup;
		`,
	})
}

// initDB initializes the database.
func (app *app) initDB() error {
	if app.dbPath == "" {
		return fmt.Errorf("database path can't be empty")
	}

	var err error
	if app.db, err = sql.Open("sqlite3", app.dbPath); err != nil {
		return fmt.Errorf("can't initialize database connection: %v", err)
	}

	return migrator.ToLatest(app.db, getDBRepo())
}

// getFileIDByID checks for a given ID in an archive, if it exists we'll
// return the fileID for that file.
func (app *app) getFileIDByID(name, id string) string {
	var fileID string
	app.queryRow(`
		SELECT file_id
		FROM archive_item
		WHERE archive = ? AND id = ? AND deleted_at IS NULL
	`, name, id).Scan(&fileID)
	return fileID
}

// getFileIDByMD5Sum checks for a given MD5sum in an archive, if it exists
// we'll return the fileID for that file.
func (app *app) getFileIDByMD5Sum(name, md5sum string) string {
	var fileID string
	app.queryRow(`
		SELECT file_id
		FROM archive_item
		WHERE archive = ? AND md5sum = ? AND deleted_at IS NULL
	`, name, md5sum).Scan(&fileID)
	return fileID
}

// addToArchive adds the item to the archive.
func (app *app) addToArchive(id, fileID, name, url, md5sum string) error {
	stmt, err := app.prepare(`
		INSERT INTO archive_item
		(id, file_id, archive, url, md5sum, created_at)
		VALUES(?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id, fileID, name, url, md5sum, newTimestamp())
	if err != nil {
		return err
	}

	return nil
}

// queryRow executes the given query and returns the result row.
func (app *app) queryRow(query string, args ...interface{}) *sql.Row {
	return app.db.QueryRow(query, args...)
}

// prepare prepares the given query and returns a statement.
func (app *app) prepare(query string) (*sql.Stmt, error) {
	return app.db.Prepare(query)
}
