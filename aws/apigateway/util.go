package apigateway

import "github.com/aws/aws-lambda-go/events"

func BuildSuccessResponse(body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: 200,
		Headers: map[string]string{
			"Access-Control-Allow-Origin" : "*",
			"Content-Type": "application/edn"},
	}
}

func BuildFailureResponse(body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: 500,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Content-Type": "application/edn"},
	}
}
