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
			name: "",
			args: args{
				subnets: []string{
					"subnet-3879e95e",
					"subnet-610b963e",
					"subnet-65456628",
					"subnet-73128d52",
				},
				svc: svc,
			},
			wantReturnAZ: map[string]string{
				"us-east-1a": "subnet-73128d52",
				"us-east-1b": "subnet-65456628",
				"us-east-1c": "subnet-610b963e",
				"us-east-1d": "subnet-3879e95e",
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
