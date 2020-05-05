package config

import (
	"testing"
)

func TestRead(t *testing.T) {
	type args struct {
		filename string
		path     string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Read db config in file",
			args: args{
				filename: "test_config",
				path:     "../tools",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Read(tt.args.filename, tt.args.path)
			if err != nil {
				t.Errorf("Read() error = %v", err)
				return
			}
		})
	}
}

func TestReadFullPath(t *testing.T) {
	type args struct {
		fullPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Read db config in file",
			args: args{
				fullPath: "../tools/test_config.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ReadFullPath(tt.args.fullPath)
			if err != nil {
				t.Errorf("ReadFullPath() error = %v", err)
				return
			}
		})
	}
}
