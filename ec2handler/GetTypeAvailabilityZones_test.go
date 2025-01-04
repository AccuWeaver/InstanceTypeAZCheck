package ec2handler

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"testing"

	"github.com/aws/aws-lambda-go/lambdacontext"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/golang/mock/gomock"
)

func TestGetTypeAvailabilityZones(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEC2Client := &MockEC2Client{}

	type args struct {
		ctx          context.Context
		instanceType string
		subnets      []string
	}
	tests := []struct {
		name                   string
		args                   args
		setup                  func()
		wantPhysicalResourceId string
		wantAzInfo             []string
		wantSubnetInfo         []string
		wantFirstSubnet        string
		wantFirstAZ            string
		wantNextIP             string
		wantErr                bool
	}{
		{
			name: "All zones",
			args: args{
				ctx: lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
					AwsRequestID:       "test-request-id",
					InvokedFunctionArn: "arn:aws:lambda:us-east-1:123456789012:function:MyFunction",
				}),
				instanceType: "t4g.small",
				subnets: []string{
					"subnet-3879e95e",
					"subnet-610b963e",
					"subnet-65456628",
					"subnet-73128d52",
				},
			},
			setup: func() {
				mockEC2Client.mockDescribeSubnets = func(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
					return &ec2.DescribeSubnetsOutput{
						Subnets: []types.Subnet{
							{
								AvailabilityZone: aws.String("us-east-1a"),
								SubnetId:         aws.String("subnet-73128d52"),
							},
							{
								AvailabilityZone: aws.String("us-east-1b"),
								SubnetId:         aws.String("subnet-65456628"),
							},
							{
								AvailabilityZone: aws.String("us-east-1c"),
								SubnetId:         aws.String("subnet-610b963e"),
							},
							{
								AvailabilityZone: aws.String("us-east-1d"),
								SubnetId:         aws.String("subnet-3879e95e"),
							},
						},
					}, nil
				}
			},
			wantPhysicalResourceId: "InstanceTypAZCheck-t4g.small",
			wantAzInfo: []string{
				"us-east-1d",
				"us-east-1a",
				"us-east-1c",
				"us-east-1b",
			},
			wantSubnetInfo: []string{
				"subnet-3879e95e",
				"subnet-610b963e",
				"subnet-65456628",
				"subnet-73128d52",
			},
			wantFirstAZ:     "us-east-1d",
			wantFirstSubnet: "subnet-3879e95e",
			wantNextIP:      "192.0.0.5",
			wantErr:         false,
		},
		{
			name: "Just 1a",
			args: args{
				ctx:          context.Background(),
				instanceType: "t4g.small",
				subnets: []string{
					"subnet-73128d52",
				},
			},
			setup: func() {
				mockEC2Client.mockDescribeSubnets = func(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
					return &ec2.DescribeSubnetsOutput{
						Subnets: []types.Subnet{
							{
								AvailabilityZone: aws.String("us-east-1a"),
								SubnetId:         aws.String("subnet-73128d52"),
							},
						},
					}, nil
				}
			},
			wantPhysicalResourceId: "InstanceTypAZCheck-t4g.small",
			wantAzInfo: []string{
				"us-east-1a",
			},
			wantSubnetInfo: []string{
				"subnet-73128d52",
			},
			wantErr: false,
		},
		{
			name: "Just 1e",
			args: args{
				ctx:          context.Background(),
				instanceType: "t4g.small",
				subnets: []string{
					"subnet-0d8f283c",
				},
			},
			setup: func() {
				mockEC2Client.mockDescribeSubnets = func(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
					return &ec2.DescribeSubnetsOutput{
						Subnets: []types.Subnet{},
					}, nil
				}
			},
			wantPhysicalResourceId: "InstanceTypAZCheck-t4g.small",
			wantAzInfo:             []string{},
			wantErr:                false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			gotPhysicalResourceId, gotAzInfo, gotSubnetInfo, gotFirstSubnet, gotFirstAZ, gotNextIP, err := GetTypeAvailabilityZones(tt.args.ctx, tt.args.instanceType, tt.args.subnets)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTypeAvailabilityZones() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPhysicalResourceId != tt.wantPhysicalResourceId {
				t.Errorf("GetTypeAvailabilityZones() gotPhysicalResourceId = %v, want %v", gotPhysicalResourceId, tt.wantPhysicalResourceId)
			}
			if !compareSlices(gotAzInfo, tt.wantAzInfo) {
				t.Errorf("GetTypeAvailabilityZones() gotAzInfo = %v, want %v", gotAzInfo, tt.wantAzInfo)
			}
			if !compareSlices(gotSubnetInfo, tt.wantSubnetInfo) {
				t.Errorf("GetTypeAvailabilityZones() gotSubnetInfo = %v, want %v", gotSubnetInfo, tt.wantSubnetInfo)
			}
			if gotFirstSubnet != tt.wantFirstSubnet {
				t.Errorf("GetTypeAvailabilityZones() gotFirstSubnet = %v, want %v", gotFirstSubnet, tt.wantFirstSubnet)
			}
			if gotFirstAZ != tt.wantFirstAZ {
				t.Errorf("GetTypeAvailabilityZones() gotFirstAZ = %v, want %v", gotFirstAZ, tt.wantFirstAZ)
			}
			if gotNextIP != tt.wantNextIP {
				t.Errorf("GetTypeAvailabilityZones() gotNextIP = %v, want %v", gotNextIP, tt.wantNextIP)
			}
		})
	}
}
