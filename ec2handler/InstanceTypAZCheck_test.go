package ec2handler

import (
	"context"
	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambdacontext"
	"testing"
)

func TestInstanceTypAZCheck(t *testing.T) {
	type args struct {
		ctx   context.Context
		event cfn.Event
	}
	tests := []struct {
		name           string
		args           args
		wantResourceId string
		wantAZInfo     map[string]interface{}
		wantErr        bool
	}{
		{
			name: "Create stack event",
			args: args{
				ctx: lambdacontext.NewContext(context.Background(), &lambdacontext.LambdaContext{
					AwsRequestID:       "test-request-id",
					InvokedFunctionArn: "arn:aws:lambda:us-east-1:123456789012:function:MyFunction",
				}),
				event: cfn.Event{
					RequestType:        "Create",
					RequestID:          "unique-id-for-request",
					ResponseURL:        "http://pre-signed-S3-url-for-response",
					ResourceType:       "Custom::InstanceTypAZCheck",
					PhysicalResourceID: "",
					LogicalResourceID:  "MyInstanceTypAZCheck",
					StackID:            "arn:aws:cloudformation:us-east-1:123456789012:stack/MyStack/guid",
					ResourceProperties: map[string]interface{}{
						"InstanceType": "t4g.small",
						"Subnets": []interface{}{
							"subnet-d17ddce0",
							"subnet-53301d1e",
							"subnet-49970916",
							"subnet-45c55823",
							"subnet-32396e3c",
							"subnet-4440d865",
						},
					},
					OldResourceProperties: nil,
				},
			},
			wantResourceId: "InstanceTypAZCheck-t4g.small",
			wantAZInfo: map[string]interface{}{
				"AvailableInAZs": []string{
					"us-east-1d",
					"us-east-1a",
					"us-east-1c",
					"us-east-1b",
					"us-east-1f",
				},
				"AvailableInSubnetIds": []string{
					"subnet-53301d1e",
					"subnet-49970916",
					"subnet-45c55823",
					"subnet-32396e3c",
					"subnet-4440d865",
				}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResourceId, gotAZInfo, err := InstanceTypAZCheck(tt.args.ctx, tt.args.event)
			if (err != nil) != tt.wantErr {
				t.Errorf("InstanceTypAZCheck() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotResourceId != tt.wantResourceId {
				t.Errorf("InstanceTypAZCheck() gotResourceId = %v, wantResourceId %v", gotResourceId, tt.wantResourceId)
			}
			// for each key in tt.wantAZInfo, check if the value is equal to the value in gotAZInfo
			for key, value := range tt.wantAZInfo {
				/// check if the key exists in gotAZInfo
				if _, ok := gotAZInfo[key]; !ok {
					t.Errorf("InstanceTypAZCheck() gotAZInfo[%v] does not exist", key)
				}
				// Now from the gotAZInfo, check make sure the value is a slice of strings and use the compareSlices
				// function to see if the slices are equal
				if gotSlice, ok := gotAZInfo[key].([]string); ok {
					if !compareSlices(gotSlice, value.([]string)) {
						t.Errorf("InstanceTypAZCheck() gotAZInfo[%v] = %v, want %v", key, gotSlice, value)
					}
				} else {
					t.Errorf("InstanceTypAZCheck() gotAZInfo[%v] is not a slice of strings", key)
				}
			}
		})
	}
}
