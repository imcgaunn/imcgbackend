package s3

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"regexp"
	"strings"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/s3"
)

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

func SplitS3Uri(s3Uri string) (bucket string, key string) {
	splitExp := regexp.MustCompile("/") // just die
	uriComponents := splitExp.Split(s3Uri, -1)
	return uriComponents[2], strings.Join(uriComponents[3:], "/")
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
