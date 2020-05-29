package actions

import (
	"fmt"
	"testing"
)

type (
	testExecutor struct {
		hasValidationError bool
		actionsValidated   []string
		actionsRun         []string
		actionResult       ActionResult
	}

	testExecutorEveryActionValid struct {
		testExecutor
	}

	testExecutorEveryActionInvalid struct {
		testExecutor
	}
)

func (t *testExecutor) Validate(action string) error {
	t.actionsValidated = append(t.actionsValidated, action)

	if t.hasValidationError {
		return fmt.Errorf("action is not valid")
	}

	return nil
}

func (t *testExecutor) Run(action string) ActionResult {
	t.actionsValidated = append(t.actionsValidated, action)

	return t.actionResult
}

func (t *testExecutor) Cleanup() {
}

func (t *testExecutorEveryActionValid) Validate(_ string) error {
	return nil
}

func (t *testExecutorEveryActionInvalid) Validate(_ string) error {
	return fmt.Errorf("action is not valid")
}

func Test_findExecutor(t *testing.T) {
	type args struct {
		action    string
		executors []Executor
	}
	tests := []struct {
		name string
		args args
		want Executor
	}{
		{
			name: "OK one executor",
			args: args{
				action: "action1",
				executors: []Executor{
					&testExecutorEveryActionValid{testExecutor{actionResult: ActionResult{Message: "executor1"}}},
				},
			},
			want: &testExecutorEveryActionValid{testExecutor{actionResult: ActionResult{Message: "executor1"}}},
		},
		{
			name: "One executor and action is not valid",
			args: args{
				action: "action1",
				executors: []Executor{
					&testExecutorEveryActionInvalid{testExecutor{actionResult: ActionResult{Message: "executor1"}}},
				},
			},
			want: nil,
		},
		{
			name: "OK two executors",
			args: args{
				action: "action1",
				executors: []Executor{
					&testExecutorEveryActionInvalid{testExecutor{actionResult: ActionResult{Message: "executor1"}}},
					&testExecutorEveryActionValid{testExecutor{actionResult: ActionResult{Message: "executor2"}}},
				},
			},
			want: &testExecutorEveryActionValid{testExecutor{actionResult: ActionResult{Message: "executor2"}}},
		},
		{
			name: "OK two executors and for both the action is valid, first executor should be returned",
			args: args{
				action: "action1",
				executors: []Executor{
					&testExecutorEveryActionValid{testExecutor{actionResult: ActionResult{Message: "executor1"}}},
					&testExecutorEveryActionValid{testExecutor{actionResult: ActionResult{Message: "executor2"}}},
				},
			},
			want: &testExecutorEveryActionValid{testExecutor{actionResult: ActionResult{Message: "executor1"}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findExecutor(tt.args.action, tt.args.executors)

			if tt.want == nil && (got != nil) {
				t.Errorf("findExecutor() = %v, want %v", got, tt.want)
			}

			if tt.want != nil && got == nil {
				t.Errorf("findExecutor() = %v, want %v", got, tt.want)
			}

			if tt.want != nil {
				if got.Run("action1").Message != tt.want.Run("action1").Message {
					t.Errorf("findExecutor() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
