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

	pickedIndex := getRandomRecordId(svc)

	result, err := svc.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key: map[string]*dynamodb.AttributeValue{
			"questionId": {
				N: aws.String(strconv.Itoa(pickedIndex)),
			},
		},
	})

	if err != nil {
		log.Printf("Error retrieving from dynamoDB", err.Error())
	}

	item := Item{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)

	mailResult, mailError := sendEmail(item)
	if mailError != nil {
		fmt.Println(mailError.Error())
	}
	fmt.Println(mailResult)
}

func getRandomRecordId(svc *dynamodb.DynamoDB) int {
	input := &dynamodb.ScanInput{
		TableName: aws.String(TableName),
		Select:    aws.String("COUNT"),
	}
	// should handle error
	scanResult, _ := svc.Scan(input)

	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	return r1.Intn(int(*scanResult.Count)) + 1
}

func sendEmail(item Item) (bool, error) {
	var HtmlBody = "<b>Question: </b><p>" + item.Question + "</p> <b>Answer:</b><p>" + item.Answer + "</p>"

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
