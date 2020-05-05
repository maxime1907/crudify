package router

import (
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/maxime1907/crudify/config"
	"github.com/maxime1907/crudify/dbhelper"
)

var configuration = config.DBInfo{
	Host:     "127.0.0.1",
	User:     "test",
	Dbname:   "test",
	Password: "testpass",
	Sslmode:  "disable",
	Driver:   "postgres",
}

func TestNew(t *testing.T) {
	dbhelper.Connect(configuration)
	tests := []struct {
		name string
	}{
		{
			name: "Initialize new router",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := New(nil, true, true)
			if router == nil {
				t.Errorf("New() = %v", router)
			}
		})
	}
}

func TestRun(t *testing.T) {
	dbhelper.Connect(configuration)
	type args struct {
		r    *mux.Router
		port int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Running router",
			args: args{
				r:    New(nil, true, true),
				port: 8080,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := make(chan string)
			go func() {
				Run(tt.args.r, tt.args.port)
				status <- "Failed to start the router"
			}()
			time.Sleep(2 * time.Second)
			select {
			case isdone := <-status:
				t.Errorf("Run() = %v", isdone)
			default:

			}
		})
	}
}
