package main

import (
	"context"
	"database/sql"
	"fmt"
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

func processIncomingPost(bucket string, key string, eventTime time.Time, downloader *s3manager.Downloader, uploader *s3manager.Uploader) {
	buffer := aws.NewWriteAtBuffer(make([]byte, 1024))
	bytesRead, err := downloader.Download(buffer, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		panic(err)
	}

	db := downloadIndexIfNecessary()
	postS3Uri := fmt.Sprintf("s3://%s/%s", bucket, key)
	ie, err := index.GetIndexEntryByS3Location(postS3Uri, db)
	if err != nil && ie.ID != 0 {
		log.Printf("there's already an index entry for this post [%s] so i'm ignoring it\n", postS3Uri)
		log.Print("the existing index entry has this info: ")
		log.Print(ie)
		return
	}

	postContent := string(buffer.Bytes()[:bytesRead])
	postLines := strings.Split(postContent, "\n")
	headerLines, err := post.ExtractPostHeaderLines(postLines)
	headerPresent := true
	if err != nil {
		log.Printf("there doesn't seem to be a real header. too bad :(")
		headerPresent = false
	}
	if headerPresent {
		postMetaData := post.ParseHeaderLines(headerLines)
		log.Print(postMetaData)
	}

	newindexEntry := index.BlogIndexEntry{PostS3Loc: postS3Uri,
		PostMetaS3Loc: "nothinyet.metadataisinline",
		CreatedTime:   eventTime}
	log.Print("adding new index entry to the database")
	res, err := index.AddIndexEntry(newindexEntry, db)
	if err != nil {
		log.Printf("failed to add index entry to the database")
		log.Fatal(err)
	}

	db.Close()
	log.Print(res)
	log.Print("added to database: ")
	log.Print(newindexEntry)
	log.Print("persisting db changes to storage backend (s3)")
	err = index.PutIndexDbFile("/tmp/index.sqlite")
	if err != nil {
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
		s3 := record.S3
		downloader := index.BuildS3DownloadManager()
		uploader := index.BuildS3UploadManager()
		processIncomingPost(s3.Bucket.Name,
			s3.Object.Key,
			evtTime,
			downloader,
			uploader)
	}
}

func main() {
	lambda.Start(updateBlogIndex)
}
