package main

import (
	"InstanceTypAZCheck/ec2handler"
	"context"
	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"log"
)

// RemoveResources - Only needed if there were actual resources created.
func RemoveResources(event cfn.Event) (*ec2.DeregisterImageOutput, error) {
	log.Printf("RemoveResources(%#v)", event)
	log.Printf("PhysicalResourceId %v", event.PhysicalResourceID)

	// Return the result
	return nil, nil
}

type Response struct {
	Status         string   `json:"Status"`
	AvailableZones []string `json:"AvailableZones"`
}

// InstanceTypAZCheck - Lambda function to get the availability zones for a given instance type
func InstanceTypAZCheck(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {
	log.Printf("InstanceTypAZCheck(%#v, %#v)", ctx, event)
	log.Printf("Context: %#v", ctx)
	log.Printf("Event: %#v", event)
	log.Printf("StackId: %#v", event.StackID)
	log.Printf("RequestType: %v", event.RequestType)
	log.Printf("ResourceType: %#v", event.ResourceType)
	log.Printf("ResourceProperties: %#v", event.ResourceProperties)
	lc, _ := lambdacontext.FromContext(ctx)

	log.Printf("AWSRequestID: %#v", lc.AwsRequestID)
	log.Printf("ClientContext: %#v", lc.ClientContext)
	log.Printf("Identity: %#v", lc.Identity)
	log.Printf("InvokedFunctionArn: %#v", lc.InvokedFunctionArn)

	// VpcId: VPC ID
	instanceType, _ := event.ResourceProperties["InstanceType"].(string)
	log.Printf("instance-type: %v", instanceType)
	subnetsInterface, _ := event.ResourceProperties["Subnets"].([]interface{})
	subnets := make([]string, len(subnetsInterface))
	for i, v := range subnetsInterface {
		subnets[i] = v.(string)
	}
	log.Printf("subnets: %v", subnets)

	physicalResourceID, typeAvailableInZones, typeAvailableInSubnetIds, err := ec2handler.GetTypeAvailabilityZones(ctx, instanceType, subnets)
	if err != nil {
		log.Printf("Error getting availability zones: %v", err)
		return "", nil, err
	}

	data := map[string]interface{}{
		"AvailableInAZs":       typeAvailableInZones,
		"AvailableInSubnetIds": typeAvailableInSubnetIds,
	}

	// Handle the different request types
	switch event.RequestType {
	case "Create":
		// No additional action needed for Create
	case "Delete":
		_, err = RemoveResources(event)
		if err != nil {
			log.Printf("Error removing resources: %v", err)
			return "", nil, err
		}
	default:
		_, err = RemoveResources(event)
		if err != nil {
			log.Printf("Error removing resources: %v", err)
			return "", nil, err
		}
		physicalResourceID, typeAvailableInZones, typeAvailableInSubnetIds, err = ec2handler.GetTypeAvailabilityZones(ctx, instanceType, subnets)
		if err != nil {
			log.Printf("Error getting availability zones: %v", err)
			return "", nil, err
		}
		data["AvailableInAZs"] = typeAvailableInZones
		data["AvailableInSubnetIds"] = typeAvailableInSubnetIds
	}

	log.Printf("Returning: %v, %#v", physicalResourceID, data)
	return physicalResourceID, data, nil
}

// main - entry point for the lambda function (wraps the actual handler so that we
// don't need to do all the boilerplate ourselves)
func main() {
	lambda.Start(cfn.LambdaWrap(InstanceTypAZCheck))
}
