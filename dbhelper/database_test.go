package dbhelper

import (
	"testing"

	"github.com/maxime1907/crudify/config"
)

const filename string = "test_config"
const path string = "../tools/"

func TestFormatSettings(t *testing.T) {
	dbInfo := config.DBInfo{
		Host:     "127.0.0.1",
		User:     "test",
		Dbname:   "test",
		Password: "testpass",
		Sslmode:  "disable",
		Driver:   "postgres",
	}

	type args struct {
		dbinfo config.DBInfo
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Format settings of struct DBInfo",
			args: args{
				dbinfo: dbInfo,
			},
			want: "host=127.0.0.1 user=test dbname=test password=testpass sslmode=disable",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatSettings(tt.args.dbinfo); got != tt.want {
				t.Errorf("formatSettings() = [%v], want [%v]", got, tt.want)
			}
		})
	}
}
