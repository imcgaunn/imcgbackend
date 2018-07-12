package index

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type BlogIndexEntry struct {
	ID            int64
	PostS3Loc     string
	PostMetaS3Loc string
	CreatedTime   time.Time
}

type BlogPost struct {
	Title   string
	Content string
}

var PostBucket string = "imcgaunn-blog-posts"
var IndexDbPath string = "index.sqlite"

func CreateIndexTable(tableName string, db *sql.DB) error {
	createStatement := `CREATE TABLE blogposts
(ID INTEGER PRIMARY KEY ASC,
 post_s3_loc text,
 post_meta_s3_loc text,
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
	insertStatement := `INSERT INTO blogposts
(post_s3_loc, post_meta_s3_loc, created_time)
VALUES(?, ?, ?)`
	stmnt, err := db.Prepare(insertStatement)
	if err != nil {
		fmt.Println(err)
		panic("failed to prepare index entry insert statement")
	}
	res, err := stmnt.Exec(e.PostS3Loc,
		e.PostMetaS3Loc,
		e.CreatedTime.Format(time.UnixDate))
	fmt.Println(res)
	return res, err
}

func RemoveIndexEntry(entryId int64, db *sql.DB) error {
	stmnt, err := db.Prepare("DELETE FROM blogposts WHERE ID=?")
	if err != nil {
		fmt.Println(err)
		panic("failed to prepare index entry delete statement")
	}
	_, err = stmnt.Exec(entryId)
	return err
}

func GetIndexEntry(entryId int64, db *sql.DB) (BlogIndexEntry, error) {
	e := &BlogIndexEntry{}
	row := db.QueryRow("SELECT * FROM blogposts WHERE ID=$1", entryId)
	err := row.Scan(&e.ID, &e.PostS3Loc, &e.PostMetaS3Loc, &e.CreatedTime)
	return *e, err
}

func GetAllIndexEntries(db *sql.DB) ([]BlogIndexEntry, error) {
	entries := []BlogIndexEntry{}
	rows, err := db.Query("SELECT * FROM blogposts")
	if err != nil {
		return entries, err
	}
	defer rows.Close()
	for rows.Next() {
		e := &BlogIndexEntry{}
		err = rows.Scan(&e.ID,
			&e.PostS3Loc,
			&e.PostMetaS3Loc,
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
	row := db.QueryRow("SELECT * from blogposts WHERE post_s3_loc=$1", location)
	err := row.Scan(&e.ID, &e.PostS3Loc, &e.PostMetaS3Loc, &e.CreatedTime)
	return *e, err
}

func FetchPostFromS3ByUri(s3Uri string) (BlogPost, error) {
	bucket, key := splitS3Uri(s3Uri)
	post, err := FetchPostFromS3(bucket, key)
	return post, err
}

func splitS3Uri(s3Uri string) (bucket string, key string) {
	splitExp := regexp.MustCompile("/") // just die
	uriComponents := splitExp.Split(s3Uri, -1)
	return uriComponents[2], strings.Join(uriComponents[3:], "/")
}

func FetchPostFromS3(bucket string, key string) (BlogPost, error) {
	stringBuilder := strings.Builder{}
	PostBytes, err := FetchBytesFromS3(bucket, key)
	if err != nil {
		return BlogPost{}, err
	}
	stringBuilder.Write(PostBytes)
	return BlogPost{
		Title:   "Great", // TODO: real code for this
		Content: stringBuilder.String(),
	}, err

}

func FetchBytesFromS3(bucket string, key string) ([]byte, error) {
	awsBuff := &aws.WriteAtBuffer{}
	downloader := BuildS3DownloadManager()
	_, err := downloader.Download(awsBuff, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return awsBuff.Bytes(), err
}

func BuildS3DownloadManager() *s3manager.Downloader {
	sess := session.Must(session.NewSession())
	svc := s3manager.NewDownloader(sess)
	return svc
}

func BuildS3UploadManager() *s3manager.Uploader {
	sess := session.Must(session.NewSession())
	svc := s3manager.NewUploader(sess)
	return svc
}

func GetIndexDbFile() ([]byte, error) {
	fmt.Println("attempting to fetch index file")
	indexFile, err := FetchBytesFromS3(PostBucket, IndexDbPath)
	if err != nil {
		fmt.Println("Failed to fetch index file :(")
	}
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
	uploader := BuildS3UploadManager()
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
