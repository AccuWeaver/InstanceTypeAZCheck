package ec2handler

import (
	"context"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"testing"
)

func TestGetTypeAvailabilityZones(t *testing.T) {
	type args struct {
		ctx          context.Context
		instanceType string
		subnets      []string
	}
	tests := []struct {
		name                   string
		args                   args
		wantPhysicalResourceId string
		wantAzInfo             []string
		wantSubnetInfo         []string
		wantErr                bool
	}{
		// TODO: Add test cases.
		{
			name: "All zones",
			args: args{
				ctx: lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
					AwsRequestID:       "test-request-id",
					InvokedFunctionArn: "arn:aws:lambda:us-east-1:123456789012:function:MyFunction",
				}),
				instanceType: "t4g.small",
				subnets: []string{
					"subnet-d17ddce0",
					"subnet-53301d1e",
					"subnet-49970916",
					"subnet-45c55823",
					"subnet-32396e3c",
					"subnet-4440d865",
				},
			},
			wantPhysicalResourceId: "InstanceTypAZCheck-t4g.small",
			wantAzInfo: []string{
				"us-east-1d",
				"us-east-1a",
				"us-east-1c",
				"us-east-1b",
				"us-east-1f",
			},
			wantSubnetInfo: []string{
				"subnet-53301d1e",
				"subnet-49970916",
				"subnet-45c55823",
				"subnet-32396e3c",
				"subnet-4440d865",
			},
			wantErr: false,
		},
		{
			name: "Just 1a",
			args: args{
				ctx:          context.Background(),
				instanceType: "t4g.small",
				subnets: []string{
					"subnet-4440d865",
				},
			},
			wantPhysicalResourceId: "InstanceTypAZCheck-t4g.small",
			wantAzInfo: []string{
				"us-east-1a",
			},
			wantSubnetInfo: []string{
				"subnet-4440d865",
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
			wantPhysicalResourceId: "InstanceTypAZCheck-t4g.small",
			wantAzInfo:             []string{},
			wantErr:                false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPhysicalResourceId, gotAzInfo, gotSubnetInfo, err := GetTypeAvailabilityZones(tt.args.ctx, tt.args.instanceType, tt.args.subnets)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTypeAvailabilityZones() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotPhysicalResourceId != tt.wantPhysicalResourceId {
				t.Errorf("GetTypeAvailabilityZones() gotPhysicalResourceId = %v, wantResourceId %v", gotPhysicalResourceId, tt.wantPhysicalResourceId)
			}

			// Compare the availability zones slices to see if they are the same length and have the same values
			if compareSlices(gotAzInfo, tt.wantAzInfo) == false {
				t.Errorf("GetTypeAvailabilityZones() gotAzInfo = %d, wantResourceId %d", len(gotAzInfo), len(tt.wantAzInfo))
				// TODO: Add a diff so we see what the differences are

			}
			if compareSlices(gotSubnetInfo, tt.wantSubnetInfo) == false {
				t.Errorf("GetTypeAvailabilityZones() gotSubnetInfo = %d, wantResourceId %d", len(gotSubnetInfo), len(tt.wantSubnetInfo))
			}
		})
	}
}
