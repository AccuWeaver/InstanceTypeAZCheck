package ec2handler

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// GetTypeAvailabilityZones - Get the availability zones for a given instance type
func GetTypeAvailabilityZones(ctx context.Context, instanceType string, subnets []string) (physicalResourceId string, availableZones []string, availableSubnets []string, err error) {
	physicalResourceId = fmt.Sprintf("InstanceTypAZCheck-%v", instanceType)
	var cfg aws.Config
	cfg, err = config.LoadDefaultConfig(ctx)
	if err != nil {
		return
	}

	svc := ec2.NewFromConfig(cfg)

	input := &ec2.DescribeInstanceTypeOfferingsInput{
		LocationType: types.LocationTypeAvailabilityZone,
		Filters: []types.Filter{
			{
				Name:   aws.String("instance-type"),
				Values: []string{instanceType},
			},
			{
				Name:   aws.String("location"),
				Values: subnets,
			},
		},
	}

	// Get the instance type offerings
	var result *ec2.DescribeInstanceTypeOfferingsOutput
	result, err = svc.DescribeInstanceTypeOfferings(ctx, input)
	if err != nil {
		return
	}
	// Get the zone IDs and add them to the list
	for _, offering := range result.InstanceTypeOfferings {
		availableZones = append(availableZones, *offering.Location)
	}

	// Describe subnets to get subnet IDs
	subnetInput := &ec2.DescribeSubnetsInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("availability-zone"),
				Values: availableZones,
			},
		},
	}

	var subnetResult *ec2.DescribeSubnetsOutput
	subnetResult, err = svc.DescribeSubnets(ctx, subnetInput)
	if err != nil {
		return
	}

	for _, subnet := range subnetResult.Subnets {
		availableSubnets = append(availableSubnets, *subnet.SubnetId)
	}

	return
}
