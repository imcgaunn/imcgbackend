package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"imcgbackend/blog/index"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
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

// TODO: figure out if it is practical to share this code with 'index' module
func BuildS3DownloadManager() *s3manager.Downloader {
	sess := session.Must(session.NewSession())
	svc := s3manager.NewDownloader(sess)
	return svc
}

// TODO: figure out if it is practical to share this code with 'index' module
func BuildS3UploadManager() *s3manager.Uploader {
	sess := session.Must(session.NewSession())
	svc := s3manager.NewUploader(sess)
	return svc
}

func extractPostHeaderLines(postLines []string) ([]string, error) {
	// fetch metadata from the beginning of the post
	// scan until you see prelude's bottom marker.
	lastHeaderRowIdx := -1
	for pos, str := range postLines {
		if str[:3] == ">>>" {
			lastHeaderRowIdx = pos
			break
		}
	}
	if lastHeaderRowIdx > 0 {
		headerLines := postLines[:lastHeaderRowIdx]
		return headerLines, nil
	}
	return nil, errors.New("failed to extract header")
}

func parseHeaderLines(lines []string) map[string]string {
	metaDataMap := make(map[string]string)
	for line := range lines {
		newMap, err := parseHeaderLine(lines[line])
		if err != nil {
			panic("encountered bad header line bye bye")
		}
		for newKey, val := range newMap {
			metaDataMap[newKey] = val
		}
	}
	return metaDataMap
}

func parseHeaderLine(line string) (map[string]string, error) {
	components := strings.Split(line, ":")
	if len(components) < 2 {
		return nil, errors.New("invalid header line")
	}
	metaData := map[string]string{
		components[0]: components[1],
	}
	return metaData, nil
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
	if err != nil {
		log.Printf("there's already an index entry for this post [%s] so i'm ignoring it\n", postS3Uri)
		log.Print("the existing index entry has this info: ")
		log.Print(ie)
		return
	}
	postContent := string(buffer.Bytes()[:bytesRead])
	postLines := strings.Split(postContent, "\n")
	headerLines, err := extractPostHeaderLines(postLines)
	if err != nil {
		log.Printf("there doesn't seem to be a real header. too bad :(")
		return
	}
	postMetaData := parseHeaderLines(headerLines)

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
	log.Print(postMetaData)

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
		downloader := BuildS3DownloadManager()
		uploader := BuildS3UploadManager()
		processIncomingPost(s3.Bucket.Name,
			s3.Object.Key,
			evtTime,
			downloader,
			uploader)
	}
	log.Printf("successfully updated index :) :) ")
}

func main() {
	lambda.Start(updateBlogIndex)
}
