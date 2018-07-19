package apigateway

import "github.com/aws/aws-lambda-go/events"

func BuildSuccessResponse(body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: 200}
}

func BuildFailureResponse(body string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: 500}
}
