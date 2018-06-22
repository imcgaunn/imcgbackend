package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/mattn/go-sqlite3"

	"imcgbackend/blog/index"
)

type Response struct {
	Message string `json:"message"`
}

func BuildSuccessResponseWithBody(body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: 200}
}

func BuildFailureResponseWithBody(body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: 500}
}

func getBlogPost(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	queryParams := request.QueryStringParameters
	postId, err := strconv.ParseInt(queryParams["id"], 10, 64)
	if err != nil {
		return BuildFailureResponseWithBody("invalid post id. must be an integer!"), err
	}
	dbFileBytes, err := index.GetIndexDbFile()
	if err != nil {
		return BuildFailureResponseWithBody("failed to get index db file"), err
	}
	err = ioutil.WriteFile("/tmp/index.sqlite", dbFileBytes, 0644)
	if err != nil {
		return BuildFailureResponseWithBody("failed to save post index"), err
	}
	dbConn, err := sql.Open("sqlite3", "file:/tmp/index.sqlite?_loc=auto")
	if err != nil {
		return BuildFailureResponseWithBody("failed access post index"), err
	}
	postIdxEntry, err := index.GetIndexEntry(postId, dbConn)
	if err != nil {
		return BuildFailureResponseWithBody(fmt.Sprintf("no post with id: %d in index", postId)), err
	}
	post, err := index.FetchPostFromS3ByUri(postIdxEntry.PostS3Loc)
	if err != nil {
		return BuildFailureResponseWithBody("d:( failed fetch )"), err
	}
	return BuildSuccessResponseWithBody(post.Content), nil
}

func main() {
	lambda.Start(getBlogPost)
}
