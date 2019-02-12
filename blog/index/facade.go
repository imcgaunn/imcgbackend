package index

import (
	"database/sql"
	"fmt"
	s3Utl "github.com/imcgaunn/imcgbackend/aws/s3"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type BlogIndexEntry struct {
	ID          int64     `edn:"id"`
	PostS3Loc   string    `edn:"post-s3-loc"`
	Title       string    `edn:"title"`
	Tags        string    `edn:"tags"`
	CreatedTime time.Time `edn:"created-time"`
}

var PostBucket string = "imcgaunn-blog-posts"
var IndexDbPath string = "index.sqlite"

func CreateIndexTable(db *sql.DB) error {
	createStatement := `CREATE TABLE blogindex
(ID INTEGER PRIMARY KEY ASC,
 post_s3_loc text,
 title text,
 tags text,
 created_time date)`
	stmnt, err := db.Prepare(createStatement)
	if err != nil {
		fmt.Println(err)
		panic("failed to prepare create table statement")
	}
	_, err = stmnt.Exec()
	return err
}

func AddIndexEntry(e BlogIndexEntry, db *sql.DB) (sql.Result, error) {
	insertStatement := `INSERT INTO blogindex
(post_s3_loc,
 title,
 tags,
 created_time)
VALUES(?, ?, ?, ?)`
	stmnt, err := db.Prepare(insertStatement)
	if err != nil {
		fmt.Println(err)
		panic("failed to prepare index entry insert statement")
	}
	res, err := stmnt.Exec(e.PostS3Loc,
		e.Title,
		e.Tags,
		e.CreatedTime.Format(time.UnixDate))
	fmt.Println(res)
	return res, err
}

func RemoveIndexEntry(entryId int64, db *sql.DB) error {
	stmnt, err := db.Prepare("DELETE FROM blogindex WHERE ID=?")
	if err != nil {
		fmt.Println(err)
		panic("failed to prepare index entry delete statement")
	}
	_, err = stmnt.Exec(entryId)
	return err
}

func GetIndexEntry(entryId int64, db *sql.DB) (BlogIndexEntry, error) {
	e := &BlogIndexEntry{}
	row := db.QueryRow("SELECT * FROM blogindex WHERE ID=$1", entryId)
	err := row.Scan(&e.ID, &e.PostS3Loc, &e.Title, &e.Tags, &e.CreatedTime)
	return *e, err
}

func GetAllIndexEntries(db *sql.DB) ([]BlogIndexEntry, error) {
	entries := []BlogIndexEntry{}
	rows, err := db.Query("SELECT * FROM blogindex")
	if err != nil {
		return entries, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &BlogIndexEntry{}
		err = rows.Scan(&e.ID,
			&e.PostS3Loc,
			&e.Title,
			&e.Tags,
			&e.CreatedTime)
		if err != nil {
			return entries, err
		}
		entries = append(entries, *e)
	}
	if err := rows.Err(); err != nil {
		return []BlogIndexEntry{}, err
	}
	return entries, nil
}

func GetIndexEntryByS3Location(location string, db *sql.DB) (BlogIndexEntry, error) {
	e := &BlogIndexEntry{}
	row := db.QueryRow("SELECT * from blogindex WHERE post_s3_loc=$1", location)
	err := row.Scan(&e.ID, &e.PostS3Loc, &e.Title, &e.Tags, &e.CreatedTime)
	return *e, err
}

func GetIndexDbFile() ([]byte, error) {
	indexFile, err := s3Utl.FetchBytesFromS3(PostBucket, IndexDbPath)
	return indexFile, err
}

func PutIndexDbFile(localPath string) error {
	fmt.Printf("attempting to persist index file to %s", localPath)
	indexBucket, indexKey := PostBucket, IndexDbPath
	indexFile, err := os.Open("/tmp/index.sqlite")
	if err != nil {
		fmt.Print(err)
		return err
	}
	uploader := s3Utl.BuildS3UploadManager()
	uploadInput := s3manager.UploadInput{
		Bucket: &indexBucket,
		Key:    &indexKey,
		Body:   indexFile,
	}
	out, err := uploader.Upload(&uploadInput)
	if err != nil {
		return err
	}
	fmt.Printf("successfully updated index in s3 at %s\n", out.Location)
	return nil
}
