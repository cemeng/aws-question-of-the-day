package main

import (
	"fmt"
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

const (
	TableName = "aws-questions"
	Sender    = "cemeng@gmail.com"
	Recipient = "cemeng@gmail.com"
	Subject   = "AWS Question"
)

type Item struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func main() {
	sess, err := session.NewSession(&aws.Config{Region: aws.String("ap-southeast-2")})
	svc := dynamodb.New(sess)
	numberOfRecords := getNumberOfRecords(svc)

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

	fmt.Println(err)

	if err != nil {
		log.Printf("Error retrieving from dynamoDB", err.Error())
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
	if mailError != nil {
		fmt.Println(mailError.Error())
	}
	fmt.Println(mailResult)
}

func getNumberOfRecords(svc *dynamodb.DynamoDB) int {
	input := &dynamodb.ScanInput{
		TableName: aws.String(TableName),
		Select:    aws.String("COUNT"),
	}
	scanResult, _ := svc.Scan(input)

	return int(*scanResult.Count)
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

	sess, err := session.NewSession(&aws.Config{Region: aws.String("us-east-1")})

	svc := ses.New(sess)

	input := &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(Recipient),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Html: &ses.Content{
					Data: aws.String(HTMLBody),
				},
			},
			Subject: &ses.Content{
				Data: aws.String(Subject),
			},
		},
		Source: aws.String(Sender),
	}

	result, err := svc.SendEmail(input)

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

	fmt.Println(result)

	return true, nil
}
