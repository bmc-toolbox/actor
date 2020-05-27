package internal

import "testing"

func Test_validateParam(t *testing.T) {
	type args struct {
		params map[string]interface{}
		param  []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "OK",
			args: args{
				params: map[string]interface{}{
					"testParam0": 0,
					"testParam1": "1",
					"testParam2": "2",
				},
				param: []string{"testParam1"},
			},
			wantErr: false,
		},
		{
			name: "OK",
			args: args{
				params: map[string]interface{}{
					"testParam0": 0,
					"testParam1": "1",
					"testParam2": "2",
				},
				param: []string{"testParam1", "testParam2"},
			},
			wantErr: false,
		},
		{
			name: "OK empty params argument",
			args: args{
				params: map[string]interface{}{
					"testParam0": 0,
					"testParam1": "1",
					"testParam2": "2",
				},
				param: []string{},
			},
			wantErr: false,
		},
		{
			name: "OK empty",
			args: args{
				params: map[string]interface{}{},
				param:  []string{},
			},
			wantErr: false,
		},
		{
			name: "Param is missed",
			args: args{
				params: map[string]interface{}{
					"testParam0": 0,
					"testParam2": "2",
				},
				param: []string{"testParam1"},
			},
			wantErr: true,
		},
		{
			name: "One param OK and another is missed",
			args: args{
				params: map[string]interface{}{
					"testParam0": 0,
					"testParam2": "2",
				},
				param: []string{"testParam1", "testParam2"},
			},
			wantErr: true,
		},
		{
			name: "Empty params map",
			args: args{
				params: map[string]interface{}{},
				param:  []string{"testParam1"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateParam(tt.args.params, tt.args.param...); (err != nil) != tt.wantErr {
				t.Errorf("validateParam() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
