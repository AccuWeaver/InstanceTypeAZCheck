package ec2handler

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"log"
	"reflect"
	"testing"
)

func TestGetSubnetDetails(t *testing.T) {
	type args struct {
		subnets []string
		svc     *ec2.Client
	}
	var cfg aws.Config
	var err error
	cfg, err = config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Printf("Error loading AWS config: %v", err)
		return
	}

	svc := ec2.NewFromConfig(cfg)
	tests := []struct {
		name         string
		args         args
		wantReturnAZ map[string]string
		wantErr      bool
	}{
		{
			name: "Subnet details",
			args: args{
				subnets: []string{
					"subnet-d17ddce0",
					"subnet-53301d1e",
					"subnet-49970916",
					"subnet-45c55823",
					"subnet-32396e3c",
					"subnet-4440d865",
				},
				svc: svc,
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
			var gotReturnAZ map[string]string
			gotReturnAZ, err = GetSubnetDetails(tt.args.subnets, tt.args.svc)
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
