package main

import (
	"InstanceTypAZCheck/ec2handler"
	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
)

// main - entry point for the lambda function
func main() {
	lambda.Start(cfn.LambdaWrap(ec2handler.InstanceTypAZCheck))
}
