package ec2handler

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambdacontext"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// GetTypeAvailabilityZones - Get the availability zones for a given instance type and subnets
func GetTypeAvailabilityZones(ctx context.Context, instanceType string, subnets []string) (physicalResourceId string, availableZones []string, availableSubnets []string, err error) {
	log.Printf("GetTypeAvailabilityZones(%#v, %v, %v)", ctx, instanceType, subnets)
	physicalResourceId = fmt.Sprintf("InstanceTypAZCheck-%v", instanceType)
	var cfg aws.Config
	cfg, err = config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Printf("Error loading AWS config: %v", err)
		return
	}

	svc := ec2.NewFromConfig(cfg)

	var azMap map[string]string
	azMap, err = GetSubnetDetails(subnets, svc)
	if err != nil {
		log.Printf("Error getting subnet details: %v", err)
		return
	}
	azKeys := make([]string, 0, len(azMap))
	for k := range azMap {
		azKeys = append(azKeys, k)
	}
	// Describe instance type offerings
	input := &ec2.DescribeInstanceTypeOfferingsInput{
		LocationType: types.LocationTypeAvailabilityZone,
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-type"),
				Values: []string{instanceType},
			},
			{
				Name:   aws.String("location"),
				Values: azKeys,
			},
		},
	}

	log.Printf("DescribeInstanceTypeOfferings input: %#v", input)

	result, err := svc.DescribeInstanceTypeOfferings(ctx, input)
	if err != nil {
		log.Printf("Error describing instance type offerings: %v", err)
		return
	}

	log.Printf("DescribeInstanceTypeOfferings result: %v", result)

	// Collect available zones
	for _, offering := range result.InstanceTypeOfferings {
		log.Printf("Adding %v to availableZones", *offering.Location)
		availableZones = append(availableZones, *offering.Location)
	}

	log.Printf("Available zones: %v", availableZones)

	// Now loop through the available zones and get the subnets
	for _, az := range availableZones {
		if subnet, ok := azMap[az]; ok {
			availableSubnets = append(availableSubnets, subnet)
		}
	}

	return
}

// GetSubnetDetails - Get the details of the subnets in the given availability zones
func GetSubnetDetails(subnets []string, svc *ec2.Client) (returnAZ map[string]string, err error) {
	returnAZ = make(map[string]string)
	var subnetDetails []types.Subnet
	var nextToken *string
	for {
		subnetInput := &ec2.DescribeSubnetsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("subnet-id"),
					Values: subnets,
				},
			},
			NextToken: nextToken,
		}

		log.Printf("DescribeSubnets input: %v", subnetInput)

		var subnetResult *ec2.DescribeSubnetsOutput
		subnetResult, err = svc.DescribeSubnets(context.Background(), subnetInput)
		if err != nil {
			log.Printf("Error describing subnets: %v", err)
			return
		}

		log.Printf("DescribeSubnets result: %v", subnetResult)

		subnetDetails = append(subnetDetails, subnetResult.Subnets...)

		if subnetResult.NextToken == nil {
			break
		}

		nextToken = subnetResult.NextToken
	}
	for _, subnet := range subnetDetails {
		returnAZ[*subnet.AvailabilityZone] = *subnet.SubnetId

	}
	return
}

// InstanceTypAZCheck - Lambda function to get the availability zones for a given instance type
func InstanceTypAZCheck(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {
	log.Printf("InstanceTypAZCheck(%#v, %#v)", ctx, event)
	lc, _ := lambdacontext.FromContext(ctx)
	log.Printf("AWSRequestID: %#v", lc.AwsRequestID)

	// Handle DELETE immediately - no resources to query or clean up
	if event.RequestType == "Delete" {
		log.Printf("DELETE event - returning success immediately")
		physicalResourceID := event.PhysicalResourceID
		if physicalResourceID == "" {
			// Fallback if PhysicalResourceID not set
			physicalResourceID = fmt.Sprintf("InstanceTypAZCheck-deleted")
		}
		return physicalResourceID, map[string]interface{}{}, nil
	}

	instanceType, ok := event.ResourceProperties["InstanceType"].(string)
	if !ok {
		err := fmt.Errorf("InstanceType property is missing or invalid")
		log.Printf("Error: %v", err)
		return "", nil, err
	}
	log.Printf("instance-type: %v", instanceType)

	subnetsInterface, ok := event.ResourceProperties["Subnets"].([]interface{})
	if !ok {
		err := fmt.Errorf("Subnets property is missing or invalid")
		log.Printf("Error: %v", err)
		return "", nil, err
	}
	subnets := make([]string, len(subnetsInterface))
	for i, v := range subnetsInterface {
		subnets[i] = v.(string)
	}
	log.Printf("subnets: %v", subnets)

	physicalResourceID, typeAvailableInZones, typeAvailableInSubnetIds, err := GetTypeAvailabilityZones(ctx, instanceType, subnets)
	if err != nil {
		log.Printf("Error getting availability zones: %v", err)
		return "", nil, err
	}

	data := map[string]interface{}{
		"AvailableInAZs":       typeAvailableInZones,
		"AvailableInSubnetIds": typeAvailableInSubnetIds,
	}

	log.Printf("Returning: %v, %#v", physicalResourceID, data)
	return physicalResourceID, data, nil
}

// compareSlices checks if two slices have the same members
func compareSlices(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	counts := make(map[string]int)

	for _, item := range slice1 {
		counts[item]++
	}

	for _, item := range slice2 {
		counts[item]--
		if counts[item] < 0 {
			return false
		}
	}

	for _, count := range counts {
		if count != 0 {
			return false
		}
	}

	return true
}
