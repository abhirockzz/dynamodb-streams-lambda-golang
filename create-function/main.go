package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func main() {
	lambda.Start(handler)
}

type User struct {
	Email   string
	State   string
	City    string
	Zipcode string
}

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

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	payload := req.Body
	log.Println("payload", payload)

	var user User
	err := json.Unmarshal([]byte(payload), &user)

	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}

	item := make(map[string]types.AttributeValue)

	item["email"] = &types.AttributeValueMemberS{Value: user.Email}
	item["state"] = &types.AttributeValueMemberS{Value: user.State}
	item["city"] = &types.AttributeValueMemberS{Value: user.City}
	item["zipcode"] = &types.AttributeValueMemberN{Value: user.Zipcode}
	item["active"] = &types.AttributeValueMemberBOOL{Value: true}

	_, err = client.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName: aws.String(table),
		Item:      item,
	})

	if err != nil {
		return events.APIGatewayV2HTTPResponse{}, err
	}

	return events.APIGatewayV2HTTPResponse{StatusCode: http.StatusCreated}, nil
}
