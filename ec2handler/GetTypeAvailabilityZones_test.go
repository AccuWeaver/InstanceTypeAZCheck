package ec2handler

import (
	"context"
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
				ctx:          context.Background(),
				instanceType: "t4g.small",
				subnets: []string{
					"us-east-1a",
					"us-east-1b",
					"us-east-1c",
					"us-east-1d",
					"us-east-1e",
				},
			},
			wantPhysicalResourceId: "InstanceTypAZCheck-t4g.small",
			wantAzInfo: []string{
				"us-east-1d",
				"us-east-1a",
				"us-east-1c",
				"us-east-1b",
			},
			wantSubnetInfo: []string{
				"subnet-53301d1e",
				"subnet-49970916",
				"subnet-45c55823",
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
					"us-east-1a",
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
					"us-east-1e",
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
				t.Errorf("GetTypeAvailabilityZones() gotPhysicalResourceId = %v, want %v", gotPhysicalResourceId, tt.wantPhysicalResourceId)
			}

			// Compare the availability zones slices to see if they are the same length and have the same values
			if compareSlices(gotAzInfo, tt.wantAzInfo) == false {
				t.Errorf("GetTypeAvailabilityZones() gotAzInfo = %d, want %d", len(gotAzInfo), len(tt.wantAzInfo))
				// TODO: Add a diff so we see what the differences are

			}
			if compareSlices(gotSubnetInfo, tt.wantSubnetInfo) == false {
				t.Errorf("GetTypeAvailabilityZones() gotSubnetInfo = %d, want %d", len(gotSubnetInfo), len(tt.wantAzInfo))
			}
		})
	}
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
