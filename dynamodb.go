package main

import (
	"errors"
	"fmt"
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
}
