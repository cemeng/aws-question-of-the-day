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

type Item struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

const (
	TableName = "aws-questions"
	Sender    = "cemeng@gmail.com"
	Recipient = "cemeng@gmail.com"
	Subject   = "AWS Question"
)

func Handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// TODO: workout cloudwatch logs
	// stdout and stderr are sent to AWS CloudWatch Logs
	// log.Printf("Processing Lambda request %s\n", request.RequestContext.RequestID)

	sess, err := session.NewSession(&aws.Config{Region: aws.String("ap-southeast-2")})
	svc := dynamodb.New(sess)

	pickedIndex := getRandomRecordId(svc)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("aws-questions"),
		Key: map[string]*dynamodb.AttributeValue{
			"questionId": {
				N: aws.String(strconv.Itoa(pickedIndex)),
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

	mailResult, mailError := sendEmail(item)
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

func getRandomRecordId(svc *dynamodb.DynamoDB) int {
	input := &dynamodb.ScanInput{
		TableName: aws.String(TableName),
		Select:    aws.String("COUNT"),
	}
	// FIXME: error is not handled
	scanResult, _ := svc.Scan(input)

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(int(*scanResult.Count)) + 1
}

func sendEmail(item Item) (bool, error) {
	// no SES service in ap-southeast-2, hence using us-east-1
	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})
	svc := ses.New(sess)
	HtmlBody := "<b>Question: </b><p>" + item.Question + "</p> <b>Answer:</b><p>" + item.Answer + "</p>"

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Data: aws.String(HtmlBody),
				},
			},
			Subject: &ses.Content{
				Data: aws.String(Subject),
			},
		},
		Source: aws.String(Sender),
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
