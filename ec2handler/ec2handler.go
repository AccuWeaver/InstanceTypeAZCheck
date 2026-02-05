package ec2handler

import (
	"context"
	"fmt"
	"log"

	"net"

	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambdacontext"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// EC2Client is an interface that defines the methods used from the ec2.Client.
type EC2Client interface {
	DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
}

// GetTypeAvailabilityZones - Get the availability zones for a given instance type and subnets
func GetTypeAvailabilityZones(ctx context.Context, instanceType string, subnets []string) (physicalResourceId string, availableZones []string, availableSubnets []string, firstSubnetId string, firstAZ string, nextIP string, err error) {
	log.Printf("GetTypeAvailabilityZones(%#v, %v, %v)", ctx, instanceType, subnets)
	lc, _ := lambdacontext.FromContext(ctx)
	physicalResourceId = fmt.Sprintf("InstanceTypAZCheck-%v-%v", instanceType, lc.AwsRequestID)
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

	// Get the first subnet ID and availability zone
	if len(availableSubnets) > 0 {
		firstSubnetId = availableSubnets[0]
	}
	if len(availableZones) > 0 {
		firstAZ = availableZones[0]
	}
	// Get the next available IP address in the first subnet
	if len(availableSubnets) > 0 {
		// describe the subnet to get the CIDR block
		subnetInput := &ec2.DescribeSubnetsInput{
			SubnetIds: []string{firstSubnetId},
		}
		var subnetDetails *ec2.DescribeSubnetsOutput
		subnetDetails, err = svc.DescribeSubnets(ctx, subnetInput)
		if err != nil {
			log.Printf("Error describing subnet: %v", err)
			return
		}
		// Get the next available IP address by using the cidr block of the subnet, and
		// stepping through the addresses until one is not in use (0-3 and 255 are reserved)
		nextIP, err = GetNextAvailableIP(*subnetDetails.Subnets[0].CidrBlock, svc)
	}
	return
}

func GetNextAvailableIP(cidrBlock string, svc *ec2.Client) (string, error) {
	// Use the cidr to get the fourth IP address
	_, ipnet, err := net.ParseCIDR(cidrBlock)
	if err != nil {
		return "", fmt.Errorf("invalid CIDR block: %v", err)
	}
	// Start checking from the .4 address
	ip := ipnet.IP
	ip[3] = 4

	for ipnet.Contains(ip) {
		// Check if the IP address is in use
		inUse, err := isIPInUse(ip.String(), svc)
		if err != nil {
			return "", err
		}
		if !inUse {
			return ip.String(), nil
		}
		// Move to the next IP address
		ip[3]++
	}

	return "", fmt.Errorf("no available IP addresses in the subnet")
}

// isIPInUse - Check if the IP address is in use
func isIPInUse(ip string, svc *ec2.Client) (bool, error) {
	// Implement the logic to check if the IP address is in use
	// This can be done by describing the network interfaces and checking the private IP addresses
	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("private-ip-address"),
				Values: []string{ip},
			},
		},
	}

	result, err := svc.DescribeNetworkInterfaces(context.Background(), input)
	if err != nil {
		return false, err
	}

	return len(result.NetworkInterfaces) > 0, nil
}

// GetSubnetDetails - Get the details of the subnets in the given availability zones
func GetSubnetDetails(subnets []string, svc EC2Client) (returnAZ map[string]string, err error) {
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

		//log.Printf("DescribeSubnets input: %v", subnetInput)

		var subnetResult *ec2.DescribeSubnetsOutput
		subnetResult, err = svc.DescribeSubnets(context.Background(), subnetInput)
		if err != nil {
			log.Printf("Error describing subnets: %v", err)
			return
		}

		subnetDetails = append(subnetDetails, subnetResult.Subnets...)

		if subnetResult.NextToken == nil {
			break
		}

		nextToken = subnetResult.NextToken
	}
	// Get the subnet IDs as a slice
	for _, subnet := range subnetDetails {
		returnAZ[*subnet.AvailabilityZone] = *subnet.SubnetId

	}

	log.Printf("Found %d subnets", len(returnAZ))

	return
}

// InstanceTypAZCheck - Lambda function to get the availability zones for a given instance type
func InstanceTypAZCheck(ctx context.Context, event cfn.Event) (string, map[string]interface{}, error) {
	log.Printf("InstanceTypAZCheck(%#v, %#v)", ctx, event)
	lc, _ := lambdacontext.FromContext(ctx)
	log.Printf("AWSRequestID: %#v", lc.AwsRequestID)
	physicalResourceID := fmt.Sprintf("InstanceTypAZCheck-%v", lc.AwsRequestID)

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
	physicalResourceID = fmt.Sprintf("InstanceTypAZCheck-%v-%v-%v", instanceType, instanceType, lc.AwsRequestID)

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

	var typeAvailableInZones []string
	var typeAvailableInSubnetIds []string
	var firstSubnetId string
	var firstAZ string
	var nextSubnetIP string
	var err error
	physicalResourceID, typeAvailableInZones, typeAvailableInSubnetIds, firstSubnetId, firstAZ, nextSubnetIP, err = GetTypeAvailabilityZones(ctx, instanceType, subnets)
	if err != nil {
		log.Printf("Error getting availability zones: %v", err)
		return "", nil, err
	}

	// Build response data - only return arrays of available zones/subnets
	data := map[string]interface{}{
		"AvailableInAZs":       typeAvailableInZones,
		"AvailableInSubnetIds": typeAvailableInSubnetIds,
		"SubnetId":             firstSubnetId,
		"AZ":                   firstAZ,
		"PrivateIP":            nextSubnetIP,
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
