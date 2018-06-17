package blog

import (
	"database/sql"
	"log"
	"time"
)

type BlogIndexEntry struct {
	ID               int64
	post_s3_loc      string
	post_meta_s3_loc string
	created_time     time.Time
}

func CreateIndexTable(tableName string, db *sql.DB) error {
	createStatement := `CREATE TABLE blogposts
(ID INTEGER PRIMARY KEY ASC,
 post_s3_loc text,
 post_meta_s3_loc text,
 created_time date)`
	stmnt, err := db.Prepare(createStatement)
	if err != nil {
		log.Print(err)
		panic("failed to prepare create table statement")
	}
	_, err = stmnt.Exec()
	return err
}

func AddIndexEntry(e BlogIndexEntry, db *sql.DB) (sql.Result, error) {
	insertStatement := `INSERT INTO blogposts
(ID, post_s3_loc, post_meta_s3_loc, created_time)
VALUES(?, ?, ?, ?)`
	stmnt, err := db.Prepare(insertStatement)
	if err != nil {
		log.Print(err)
		panic("failed to prepare index entry insert statement")
	}
	res, err := stmnt.Exec(e.ID,
		e.post_s3_loc,
		e.post_meta_s3_loc,
		e.created_time.Format(time.UnixDate))
	log.Print(res)
	return res, err
}

func RemoveIndexEntry(entryId int64, db *sql.DB) error {
	stmnt, err := db.Prepare("DELETE FROM blogposts WHERE ID=?")
	if err != nil {
		log.Print(err)
		panic("failed to prepare index entry delete statement")
	}
	_, err = stmnt.Exec(entryId)
	return err
}

func GetIndexEntry(entryId int64, db *sql.DB) (BlogIndexEntry, error) {
	var e BlogIndexEntry
	row := db.QueryRow("SELECT * FROM blogposts WHERE ID=$1", entryId)
	err := row.Scan(&e.ID, &e.post_s3_loc, &e.post_meta_s3_loc, &e.created_time)
	return e, err
}

func FetchPost(s3Uri string) error {
	// go out to s3 and fetch the blog post bytes.
	return nil
}

func FetchPostMeta(s3Uri string) error {
	// go out to s3 and fetch the blog post metadata bytes.
	return nil
}
