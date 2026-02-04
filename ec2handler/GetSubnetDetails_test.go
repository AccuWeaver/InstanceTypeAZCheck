package ec2handler

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/golang/mock/gomock"
	"reflect"
	"testing"
)

// MockEC2Client is a mock implementation of the EC2Client interface.
type MockEC2Client struct {
	mockDescribeSubnets func(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
}

func (m *MockEC2Client) DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
	return m.mockDescribeSubnets(ctx, params, optFns...)
}

func TestGetSubnetDetails(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockEC2Client := &MockEC2Client{}
	type args struct {
		subnets []string
		svc     EC2Client
	}

	tests := []struct {
		name         string
		args         args
		setup        func()
		wantReturnAZ map[string]string
		wantErr      bool
	}{
		{
			name: "Valid subnets",
			args: args{
				subnets: []string{
					"subnet-d17ddce0",
					"subnet-53301d1e",
					"subnet-49970916",
					"subnet-45c55823",
					"subnet-32396e3c",
					"subnet-4440d865",
				},
				svc: mockEC2Client,
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
			wantReturnAZ: map[string]string{
				"us-east-1a": "subnet-4440d865",
				"us-east-1b": "subnet-53301d1e",
				"us-east-1c": "subnet-49970916",
				"us-east-1d": "subnet-45c55823",
				"us-east-1e": "subnet-d17ddce0",
				"us-east-1f": "subnet-32396e3c",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			gotReturnAZ, err := GetSubnetDetails(tt.args.subnets, tt.args.svc)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSubnetDetails() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotReturnAZ, tt.wantReturnAZ) {
				t.Errorf("GetSubnetDetails() gotReturnAZ = %v, want %v", gotReturnAZ, tt.wantReturnAZ)
			}
		})
	}
}
