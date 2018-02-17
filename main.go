package main

import (
	"errors"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"log"
)

type Item struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// stdout and stderr are sent to AWS CloudWatch Logs
	// log.Printf("Processing Lambda request %s\n", request.RequestContext.RequestID)

	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("ap-southeast-2")},
	)
	svc := dynamodb.New(sess)
	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("aws-questions"),
		Key: map[string]*dynamodb.AttributeValue{
			"questionId": {
				N: aws.String("1"),
			},
		},
	})

	if err != nil {
		log.Printf("Error retrieving from dynamoDB", err.Error())
		var ErrRetrievingItem = errors.New("Error retrieving item " + err.Error())
		return events.APIGatewayProxyResponse{}, ErrRetrievingItem
	}

	item := Item{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	return events.APIGatewayProxyResponse{
		Body:       "Question: " + item.Question + ", Answer: " + item.Answer,
		StatusCode: 200,
	}, nil

}

func main() {
	lambda.Start(Handler)
}
