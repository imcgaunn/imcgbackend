package main

import (
	"context"
	"database/sql"
	"imcgbackend/aws/apigateway"
	"imcgbackend/blog/index"
	"io/ioutil"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	_ "github.com/mattn/go-sqlite3"
	"olympos.io/encoding/edn"
)

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

func GetBlogIndex(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	conn := downloadIndexIfNecessary()
	indexEntries, err := index.GetAllIndexEntries(conn)
	if err != nil {
		return apigateway.BuildFailureResponse("failed to retrieve index entries", map[string]string{
			"Access-Control-Allow-Origin" : "*",
			"Content-Type": "text/plain"}), err
	}
	indexEntriesEdn, err := edn.MarshalIndent(indexEntries, "", "  ")
	return apigateway.BuildSuccessResponse(string(indexEntriesEdn), map[string]string{
		"Access-Control-Allow-Origin" : "*",
		"Content-Type": "application/edn"}), nil
}

func main() {
	lambda.Start(GetBlogIndex)
}
