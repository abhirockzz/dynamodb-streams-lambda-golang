package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var table string
var client *dynamodb.Client

func init() {
	table = os.Getenv("TABLE_NAME")
	if table == "" {
		log.Fatal("missing environment variable TABLE_NAME")
	}
	cfg, _ := config.LoadDefaultConfig(context.Background())
	client = dynamodb.NewFromConfig(cfg)

}

func main() {
	lambda.Start(handler)
}

func handler(ctx context.Context, e events.DynamoDBEvent) {

	for _, r := range e.Records {
		log.Println("New record -", r.Change.NewImage)

		item := make(map[string]types.AttributeValue)

		for k, v := range r.Change.NewImage {
			log.Println("arrtibute info", k, v, v.DataType())
			if v.DataType() == events.DataTypeString {
				item[k] = &types.AttributeValueMemberS{Value: v.String()}
			} else if v.DataType() == events.DataTypeBoolean {
				item[k] = &types.AttributeValueMemberBOOL{Value: v.Boolean()}
			}
		}

		_, err := client.PutItem(context.Background(), &dynamodb.PutItemInput{
			TableName: aws.String(table),
			Item:      item})

		if err != nil {
			log.Println(err)
		}
	}
}
