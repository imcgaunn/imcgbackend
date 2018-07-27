package main

import (
	"context"
	"database/sql"
	"fmt"
	s3Utl "imcgbackend/aws/s3"
	"imcgbackend/blog/index"
	"imcgbackend/blog_indexer/post"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	_ "github.com/mattn/go-sqlite3"
)

func printEventDetails(eventRecord events.S3EventRecord) {
	log.Printf("--\n")
	log.Printf("event name: [%s]\n", eventRecord.EventName)
	log.Printf("event source: [%s]\n", eventRecord.EventSource)
	log.Printf("event time: [%s]\n", eventRecord.EventTime)
	log.Printf("bucket, key: [%s, %s]\n", eventRecord.S3.Bucket.Name, eventRecord.S3.Object.Key)
	log.Printf("--\n")
}

func downloadIndexIfNecessary() *sql.DB {
	// TODO: more conditional love
	dbBytes, err := index.GetIndexDbFile()
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile("/tmp/index.sqlite", dbBytes, 0644)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := sql.Open("sqlite3", "file:/tmp/index.sqlite?_loc=auto")
	if err != nil {
		log.Fatal(err)
	}
	return conn
}

func FetchPostString(downloader *s3manager.Downloader, bucket string, key string) string {
	buffer := aws.NewWriteAtBuffer(make([]byte, 1024))
	bytesRead, err := downloader.Download(buffer, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		panic(err)
	}
	postContent := string(buffer.Bytes()[:bytesRead])
	return postContent
}

func existingIndexEntry(postUri string, conn *sql.DB) bool {
	index.GetIndexEntryByS3Location(postUri, conn)
	ie, err := index.GetIndexEntryByS3Location(postUri, conn)
	if err != nil && ie.ID != 0 {
		log.Printf("there's already an index entry for this post [%s]\n", postUri)
		log.Print("the existing index entry has this info: ")
		log.Print(ie)
		return true
	}
	return false
}

func rebuildPostBody(postLines []string) string {
	postBodyBuilder := strings.Builder{}
	for _, str := range postLines {
		postBodyBuilder.WriteString(str)
	}
	postBodyString := postBodyBuilder.String()
	log.Printf("compiled post body after header as single string: %s\n\n", postBodyString)
	return postBodyString
}

func persistChangesToStorage(indexDbPath string, postContent string, bucket string, key string, uploader *s3manager.Uploader) error {
	log.Print("persisting db changes to storage backend (s3)")
	err := index.PutIndexDbFile(indexDbPath)
	if err != nil {
		log.Printf("failed to persist index file to s3\n")
		return err
	}
	uploadParams := &s3manager.UploadInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   strings.NewReader(postContent)}
	_, err = uploader.Upload(uploadParams)
	return err
}

func processIncomingPost(bucket string, key string, eventTime time.Time, downloader *s3manager.Downloader, uploader *s3manager.Uploader) {

	postContent := FetchPostString(downloader, bucket, key)

	db := downloadIndexIfNecessary()
	postS3Uri := fmt.Sprintf("s3://%s/%s", bucket, key)
	if existingIndexEntry(postS3Uri, db) {
		return
	}

	postLines := strings.Split(postContent, "\n")
	headerLines, headerEndIdx, err := post.ExtractPostHeaderLines(postLines)
	headerPresent := true
	if err != nil {
		log.Printf("there doesn't seem to be a real header. too bad :(")
		log.Print(err)
		headerPresent = false
	}
	if !headerPresent {
		log.Printf("nothing to index if there's no header")
		return
	}

	postMetaData := post.ParseHeaderLines(headerLines)
	log.Print(postMetaData)
	postTitle := postMetaData["title"]
	postTags := postMetaData["tags"]
	if postTitle == "" {
		log.Fatal("missing required metadata attribute 'title'")
	}
	linesAfterHeader := postLines[headerEndIdx+1:]
	postString := rebuildPostBody(linesAfterHeader)
	newIndexEntry := index.BlogIndexEntry{PostS3Loc: postS3Uri,
		Title:       postTitle,
		Tags:        postTags,
		CreatedTime: eventTime}

	log.Print("adding new index entry to the database")
	res, err := index.AddIndexEntry(newIndexEntry, db)
	if err != nil {
		log.Printf("failed to add index entry to the database\n")
		log.Fatal(err)
	}
	log.Print(res)
	log.Print("added to database: ")
	log.Print(newIndexEntry)
	db.Close()

	err = persistChangesToStorage("/tmp/index.sqlite",
		postString,
		bucket,
		key,
		uploader)
	if err != nil {
		log.Printf("failed to persist indexer changes to backend storage")
		log.Fatal(err)
	}
}

func updateBlogIndex(ctx context.Context, s3Event events.S3Event) {
	log.Printf("we are about to update the index! wish me luck")
	for _, record := range s3Event.Records {
		printEventDetails(record)

		// bail out if this isn't a putObject event
		// no test events allowed!
		if record.EventName != "ObjectCreated:Put" {
			return
		}

		evtTime := record.EventTime
		s3Obj := record.S3

		downloader := s3Utl.BuildS3DownloadManager()
		uploader := s3Utl.BuildS3UploadManager()
		processIncomingPost(s3Obj.Bucket.Name,
			s3Obj.Object.Key,
			evtTime,
			downloader,
			uploader)
	}
}

func main() {
	lambda.Start(updateBlogIndex)
}
