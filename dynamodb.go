package main

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/ses"
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

func main() {
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
	}

	item := Item{}
	err = dynamodbattribute.UnmarshalMap(result.Item, &item)
	fmt.Println(item.Question)
	fmt.Println(item.Answer)

	mailResult, mailError := sendEmail(item)
	if mailError != nil {
		fmt.Println(mailError.Error())
	}
	fmt.Println(mailResult)
}

func sendEmail(item Item) (bool, error) {
	const (
		Sender    = "cemeng@gmail.com"
		Recipient = "cemeng@gmail.com"
		Subject   = "AWS Question"
	)

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
