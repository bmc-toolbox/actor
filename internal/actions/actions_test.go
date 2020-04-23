package actions

import (
	"testing"
	"time"
)

func Test_parserDuration(t *testing.T) {
	type args struct {
		sleepAction string
	}
	tests := []struct {
		name    string
		args    args
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "1s",
			args:    args{"sleep 1s"},
			want:    1 * time.Second,
			wantErr: false,
		},
		{
			name:    "10m",
			args:    args{"sleep 10m"},
			want:    10 * time.Minute,
			wantErr: false,
		},
		{
			name:    "invalid duration: missing unit",
			args:    args{"sleep 10"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid duration: missing value",
			args:    args{"sleep s"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid action: sleep has prefix",
			args:    args{"asleep 10s"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid action: sleep has postfix",
			args:    args{"sleepz 10s"},
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty",
			args:    args{""},
			want:    0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parserDuration(tt.args.sleepAction)
			if (err != nil) != tt.wantErr {
				t.Errorf("parserDuration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parserDuration() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsSleepAction(t *testing.T) {
	type args struct {
		action string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "ok",
			args:    args{"sleep 10s"},
			want:    true,
			wantErr: false,
		},
		{
			name:    "empty",
			args:    args{""},
			want:    false,
			wantErr: false,
		},
		{
			name:    "'sleep' has prefix",
			args:    args{"asleep 10s"},
			want:    false,
			wantErr: false,
		},
		{
			name:    "'sleep' has postfix",
			args:    args{"sleepz 10s"},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsSleepAction(tt.args.action)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSleepAction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsSleepAction() got = %v, want %v", got, tt.want)
			}
		})
	}
}
