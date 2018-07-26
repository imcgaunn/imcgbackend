package apigateway

import "github.com/aws/aws-lambda-go/events"

func BuildSuccessResponse(body string, headers map[string]string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: 200,
		Headers: headers,
	}
}

func BuildFailureResponse(body string, headers map[string]string) events.APIGatewayProxyResponse {
	return events.APIGatewayProxyResponse{
		Body:       body,
		StatusCode: 500,
		Headers: headers,
	}
}
