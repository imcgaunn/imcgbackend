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

	"imcgbackend/aws/apigateway"
	"imcgbackend/blog/index"
)

func GetBlogPost(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	queryParams := request.QueryStringParameters
	postID, err := strconv.ParseInt(queryParams["id"], 10, 64)
	if err != nil {
		return apigateway.BuildFailureResponse("invalid post id. must be an integer!"), err
	}
	dbFileBytes, err := index.GetIndexDbFile()
	if err != nil {
		return apigateway.BuildFailureResponse("failed to get index db file"), err
	}
	err = ioutil.WriteFile("/tmp/index.sqlite", dbFileBytes, 0644)
	if err != nil {
		return apigateway.BuildFailureResponse("failed to save post index"), err
	}
	dbConn, err := sql.Open("sqlite3", "file:/tmp/index.sqlite?_loc=auto")
	if err != nil {
		return apigateway.BuildFailureResponse("failed access post index"), err
	}
	postIdxEntry, err := index.GetIndexEntry(postID, dbConn)
	if err != nil {
		return apigateway.BuildFailureResponse(fmt.Sprintf("no post with id: %d in index", postID)), err
	}
	post, err := index.FetchPostFromS3ByUri(postIdxEntry.PostS3Loc)
	if err != nil {
		return apigateway.BuildFailureResponse("d:( failed fetch )"), err
	}
	return apigateway.BuildSuccessResponse(post.Content), nil
}

func main() {
	lambda.Start(GetBlogPost)
}
