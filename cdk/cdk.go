package main

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdkapigatewayv2alpha/v2"
	"github.com/aws/aws-cdk-go/awscdkapigatewayv2integrationsalpha/v2"
	"github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"

	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

const envVarName = "TABLE_NAME"

const createFunctionDir = "../create-function"
const replicateFunctionDir = "../replicate-function"

//const sourceTableName = "users"
//const targetTableName = "users_by_state"

type DynamoDBStreamsLambdaGolangStackProps struct {
	awscdk.StackProps
}

func NewDynamoDBStreamsLambdaGolangStack(scope constructs.Construct, id string, props *DynamoDBStreamsLambdaGolangStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	sourceDynamoDBTable := awsdynamodb.NewTable(stack, jsii.String("source-dynamodb-table"),
		&awsdynamodb.TableProps{
			PartitionKey: &awsdynamodb.Attribute{
				Name: jsii.String("email"),
				Type: awsdynamodb.AttributeType_STRING},
			Stream: awsdynamodb.StreamViewType_NEW_AND_OLD_IMAGES})

	sourceDynamoDBTable.ApplyRemovalPolicy(awscdk.RemovalPolicy_DESTROY)

	createUserFunction := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("create-function"),
		&awscdklambdagoalpha.GoFunctionProps{
			Runtime:     awslambda.Runtime_GO_1_X(),
			Environment: &map[string]*string{envVarName: sourceDynamoDBTable.TableName()},
			Entry:       jsii.String(createFunctionDir)})

	sourceDynamoDBTable.GrantWriteData(createUserFunction)

	api := awscdkapigatewayv2alpha.NewHttpApi(stack, jsii.String("http-api"), nil)

	createFunctionIntg := awscdkapigatewayv2integrationsalpha.NewHttpLambdaIntegration(jsii.String("create-function-integration"), createUserFunction, nil)

	api.AddRoutes(&awscdkapigatewayv2alpha.AddRoutesOptions{
		Path:        jsii.String("/"),
		Methods:     &[]awscdkapigatewayv2alpha.HttpMethod{awscdkapigatewayv2alpha.HttpMethod_POST},
		Integration: createFunctionIntg})

	targetDynamoDBTable := awsdynamodb.NewTable(stack, jsii.String("target-dynamodb-table"),
		&awsdynamodb.TableProps{
			PartitionKey: &awsdynamodb.Attribute{
				Name: jsii.String("state"),
				Type: awsdynamodb.AttributeType_STRING},
			SortKey: &awsdynamodb.Attribute{
				Name: jsii.String("city"),
				Type: awsdynamodb.AttributeType_STRING},
		})

	targetDynamoDBTable.ApplyRemovalPolicy(awscdk.RemovalPolicy_DESTROY)

	replicateUserFunction := awscdklambdagoalpha.NewGoFunction(stack, jsii.String("replicate-function"),
		&awscdklambdagoalpha.GoFunctionProps{
			Runtime:     awslambda.Runtime_GO_1_X(),
			Environment: &map[string]*string{envVarName: targetDynamoDBTable.TableName()},
			Entry:       jsii.String(replicateFunctionDir)})

	replicateUserFunction.AddEventSource(awslambdaeventsources.NewDynamoEventSource(sourceDynamoDBTable, &awslambdaeventsources.DynamoEventSourceProps{StartingPosition: awslambda.StartingPosition_LATEST, Enabled: jsii.Bool(true)}))

	targetDynamoDBTable.GrantWriteData(replicateUserFunction)

	awscdk.NewCfnOutput(stack, jsii.String("api-gateway-endpoint"),
		&awscdk.CfnOutputProps{
			ExportName: jsii.String("API-Gateway-Endpoint"),
			Value:      api.Url()})

	return stack
}

func main() {
	app := awscdk.NewApp(nil)

	NewDynamoDBStreamsLambdaGolangStack(app, "DynamoDBStreamsLambdaGolangStack", &DynamoDBStreamsLambdaGolangStackProps{
		awscdk.StackProps{
			Env: env(),
		},
	})

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	return nil
}
