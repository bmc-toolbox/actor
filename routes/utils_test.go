package routes

import (
	"testing"
)

func Test_validateBladePos(t *testing.T) {
	type args struct {
		pos string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "OK",
			args:    args{pos: "123"},
			wantErr: false,
		},
		{
			name:    "OK negative",
			args:    args{pos: "-123"},
			wantErr: false,
		},
		{
			name:    "Not a number",
			args:    args{pos: "one-two-three"},
			wantErr: true,
		},
		{
			name:    "Empty",
			args:    args{pos: ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateBladePos(tt.args.pos); (err != nil) != tt.wantErr {
				t.Errorf("validateBladePos() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateBladeSerial(t *testing.T) {
	type args struct {
		serial string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "OK",
			args:    args{serial: "qwe123"},
			wantErr: false,
		},
		{
			name:    "Empty",
			args:    args{serial: ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateBladeSerial(tt.args.serial); (err != nil) != tt.wantErr {
				t.Errorf("validateBladeSerial() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateHost(t *testing.T) {
	type args struct {
		host string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "OK fqdn",
			args:    args{host: "host.example.com"},
			wantErr: false,
		},
		{
			name:    "OK IPv4",
			args:    args{host: "1.1.1.1"},
			wantErr: false,
		},
		{
			name:    "Empty",
			args:    args{host: ""},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateHost(tt.args.host); (err != nil) != tt.wantErr {
				t.Errorf("validateHost() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
