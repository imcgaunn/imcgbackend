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

	"github.com/imcgaunn/imcgbackend/aws/apigateway"
	"github.com/imcgaunn/imcgbackend/blog/index"
	"github.com/imcgaunn/imcgbackend/blog/post"
)

func GetBlogPost(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	queryParams := request.QueryStringParameters
	postID, err := strconv.ParseInt(queryParams["id"], 10, 64)
	if err != nil {
		return apigateway.BuildFailureResponse("invalid post id. must be an integer!",
			map[string]string{
				"Access-Control-Allow-Origin" : "*",
				"Content-Type": "text/plain"}), err
	}
	dbFileBytes, err := index.GetIndexDbFile()
	if err != nil {
		return apigateway.BuildFailureResponse("failed to get index db file", map[string]string{
			"Access-Control-Allow-Origin" : "*",
			"Content-Type": "text/plain"}), err
	}
	err = ioutil.WriteFile("/tmp/index.sqlite", dbFileBytes, 0644)
	if err != nil {
		return apigateway.BuildFailureResponse("failed to save post index", map[string]string{
			"Access-Control-Allow-Origin" : "*",
			"Content-Type": "text/plain"}), err
	}
	dbConn, err := sql.Open("sqlite3", "file:/tmp/index.sqlite?_loc=auto")
	if err != nil {
		return apigateway.BuildFailureResponse("failed access post index", map[string]string{
			"Access-Control-Allow-Origin" : "*",
			"Content-Type": "text/plain"}), err
	}
	postIdxEntry, err := index.GetIndexEntry(postID, dbConn)
	if err != nil {
		return apigateway.BuildFailureResponse(fmt.Sprintf("no post with id: %d in index", postID), map[string]string{
			"Access-Control-Allow-Origin" : "*",
			"Content-Type": "text/plain"}), err
	}
	blogpost, err := post.FetchPostFromS3ByUri(postIdxEntry.PostS3Loc)
	if err != nil {
		return apigateway.BuildFailureResponse("d:( failed fetch )", map[string]string{
			"Access-Control-Allow-Origin" : "*",
			"Content-Type": "text/plain"}), err
	}
	return apigateway.BuildSuccessResponse(blogpost.Content, map[string]string{
		"Access-Control-Allow-Origin" : "*",
		"Content-Type": "text/plain"}), nil
}

func main() {
	lambda.Start(GetBlogPost)
}
