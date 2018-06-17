package blog

import (
	"database/sql"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
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

func FetchFromS3ByUri(s3Uri string) ([]byte, error) {
	// go out to s3 and fetch the blog post bytes.
	bucket, key := splitS3Uri(s3Uri)
	postBytes, err := FetchFromS3(bucket, key)
	log.Print(postBytes)
	return postBytes, err
}

func splitS3Uri(s3Uri string) (bucket string, key string) {
	splitExp := regexp.MustCompile("/")
	uriComponents := splitExp.Split(s3Uri, -1)
	return uriComponents[2], strings.Join(uriComponents[3:], "/")

}

func FetchFromS3(bucket string, key string) ([]byte, error) {
	awsBuff := aws.NewWriteAtBuffer(make([]byte, 4096)) // allocate 4k by default
	downloader := BuildS3DownloadManager()
	numBytes, err := downloader.Download(awsBuff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	log.Print(numBytes)
	return awsBuff.Bytes(), err
}

func BuildS3DownloadManager() *s3manager.Downloader {
	sess := session.Must(session.NewSession())
	svc := s3manager.NewDownloader(sess)
	return svc
}
