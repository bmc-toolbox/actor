package routes

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/bmc-toolbox/actor/internal/actions"
)

func Test_actionResultsToResponses(t *testing.T) {
	type args struct {
		results []actions.ActionResult
	}
	tests := []struct {
		name string
		args args
		want []response
	}{
		{
			name: "OK",
			args: args{
				results: []actions.ActionResult{
					{
						Action:  "test",
						Status:  false,
						Message: "ok",
						Error:   nil,
					},
					{
						Action:  "test1",
						Status:  true,
						Message: "failed",
						Error:   fmt.Errorf("test error"),
					},
				},
			},
			want: []response{
				{
					Action:  "test",
					Status:  false,
					Message: "ok",
					Error:   "",
				},
				{
					Action:  "test1",
					Status:  true,
					Message: "failed",
					Error:   "test error",
				},
			},
		},
		{
			name: "Empty",
			args: args{results: []actions.ActionResult{}},
			want: []response{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := actionResultsToResponses(tt.args.results); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("actionResultsToResponses() = %v, want %v", got, tt.want)
			}
		})
	}
}
