package post

import (
	"strings"
	s3Utl "github.com/imcgaunn/imcgbackend/aws/s3"
)


type BlogPost struct {
	Content string
}

func FetchPostFromS3ByUri(s3Uri string) (BlogPost, error) {
	bucket, key := s3Utl.SplitS3Uri(s3Uri)
	post, err := FetchPostFromS3(bucket, key)
	return post, err
}

func FetchPostFromS3(bucket string, key string) (BlogPost, error) {
	stringBuilder := strings.Builder{}
	PostBytes, err := s3Utl.FetchBytesFromS3(bucket, key)
	if err != nil {
		return BlogPost{}, err
	}
	stringBuilder.Write(PostBytes)
	return BlogPost{
		Content: stringBuilder.String(),
	}, err
}
