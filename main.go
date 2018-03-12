package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/ses"
	"log"
	"math/rand"
	"strconv"
	"time"
)

// Item is a combination of Question and Answer
type Item struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

const (
	tableName = "aws-questions"
	sender    = "cemeng@gmail.com"
	recipient = "cemeng@gmail.com"
	subject   = "AWS Question"
)

// Handler function for lambda
func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: workout cloudwatch logs
	// stdout and stderr are sent to AWS CloudWatch Logs
	// log.Printf("Processing Lambda request %s\n", request.RequestContext.RequestID)

	sess, err := session.NewSession(&aws.Config{Region: aws.String("ap-southeast-2")})
	svc := dynamodb.New(sess)
	numberOfRecords, numberOfRecordsErr := getNumberOfRecords(svc)

	if numberOfRecordsErr != nil {
		fmt.Println(numberOfRecordsErr.Error())
		return events.APIGatewayProxyResponse{
			Body:       numberOfRecordsErr.Error(),
			StatusCode: 500,
		}, nil
	}

	input := &dynamodb.BatchGetItemInput{
		RequestItems: map[string]*dynamodb.KeysAndAttributes{
			"aws-questions": {
				Keys: []map[string]*dynamodb.AttributeValue{
					{
						"questionId": &dynamodb.AttributeValue{
							N: aws.String(strconv.Itoa(getRandomRecordID(numberOfRecords))),
						},
					},
					{
						"questionId": &dynamodb.AttributeValue{
							N: aws.String(strconv.Itoa(getRandomRecordID(numberOfRecords))),
						},
					},
				},
				ProjectionExpression: aws.String("question, answer"),
			},
		},
	}
	result, err := svc.BatchGetItem(input)

	if err != nil {
		log.Printf("Error retrieving from dynamoDB", err.Error())
		var ErrRetrievingItem = errors.New("Error retrieving item " + err.Error())
		return events.APIGatewayProxyResponse{}, ErrRetrievingItem
	}

	var items [2]Item
	for index, element := range result.Responses["aws-questions"] {
		item := Item{}
		err = dynamodbattribute.UnmarshalMap(element, &item)
		if err == nil {
			items[index] = item
		} else {
			fmt.Println(err)
		}
	}

	mailResult, mailError := sendEmail(items)
	// FIXME: use mailResult
	fmt.Println(mailResult)
	if mailError != nil {
		fmt.Println(mailError.Error())
		return events.APIGatewayProxyResponse{
			Body:       mailError.Error(),
			StatusCode: 500,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		Body:       "Mail sent",
		StatusCode: 200,
	}, nil

}

func getNumberOfRecords(svc *dynamodb.DynamoDB) (int, error) {
	input := &dynamodb.ScanInput{
		TableName: aws.String(tableName),
		Select:    aws.String("COUNT"),
	}
	scanResult, err := svc.Scan(input)

	if err != nil {
		return 0, err
	}
	return int(*scanResult.Count), nil
}

func getRandomRecordID(numberOfRecords int) int {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(numberOfRecords) + 1
}

func sendEmail(items [2]Item) (bool, error) {
	HTMLBody := ""
	for _, item := range items {
		HTMLBody += "<b>Question: </b><p>" + item.Question + "</p> <b>Answer:</b><p>" + item.Answer + "</p><p>====</p>"
	}

	// no SES service in ap-southeast-2, hence using us-east-1
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	svc := ses.New(sess)

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Data: aws.String(HTMLBody),
				},
			},
			Subject: &ses.Content{
				Data: aws.String(subject),
			},
		},
		Source: aws.String(sender),
	}

	result, err := svc.SendEmail(input)
	// FIXME: return result rather than boolean
	fmt.Println(result)

	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case ses.ErrCodeMessageRejected:
				fmt.Println(ses.ErrCodeMessageRejected, aerr.Error())
			case ses.ErrCodeMailFromDomainNotVerifiedException:
				fmt.Println(ses.ErrCodeMailFromDomainNotVerifiedException, aerr.Error())
			case ses.ErrCodeConfigurationSetDoesNotExistException:
				fmt.Println(ses.ErrCodeConfigurationSetDoesNotExistException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			fmt.Println(err.Error())
		}
		return false, err
	}

	return true, nil
}

func main() {
	lambda.Start(Handler)
}
