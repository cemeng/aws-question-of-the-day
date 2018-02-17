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

var (
	ErrNameNotProvided = errors.New("no name was provided in the HTTP body")
	ErrRetrievingItem  = errors.New("error retrieving item")
)

type Item struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {

	// stdout and stderr are sent to AWS CloudWatch Logs
	log.Printf("Processing Lambda request %s\n", request.RequestContext.RequestID)

	// If no name is provided in the HTTP request body, throw an error
	if len(request.Body) < 1 {
		return events.APIGatewayProxyResponse{}, ErrNameNotProvided
	}

	// Do a dynamoDB look up
	// return on the body result
	// Python code
	// dynamodb = boto3.resource('dynamodb')
	// table = dynamodb.Table(os.environ['DB_TABLE_NAME'])
	// items = table.scan()
	// items = table.query(
	//     KeyConditionExpression=Key('id').eq(postId)
	// )
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
		return events.APIGatewayProxyResponse{}, ErrRetrievingItem
	}
	item := Item{}

	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	return events.APIGatewayProxyResponse{
		Body:       "Hello " + request.Body,
		StatusCode: 200,
	}, nil

}

func main() {
	lambda.Start(Handler)
}
