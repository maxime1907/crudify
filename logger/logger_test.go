package logger

import (
	"net/http"
	"testing"
	"time"
)

type TestRoute struct {
	Name        string
	HandlerFunc http.HandlerFunc
}

// Dummy panic handler
func testPanicFuncHandler(w http.ResponseWriter, r *http.Request, name string, start time.Time) {

}

// Dummy handler
func testHandler(w http.ResponseWriter, r *http.Request) {

}

func TestLogger(t *testing.T) {
	var route = TestRoute{
		Name:        "/test/",
		HandlerFunc: testHandler,
	}

	type args struct {
		handler   http.Handler
		name      string
		funcdefer DeferFunc
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Logger writing",
			args: args{
				handler: route.HandlerFunc,
				name:    route.Name,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Logger(tt.args.handler, tt.args.name, testPanicFuncHandler)
			if got == nil {
				t.Errorf("Logger() = %v", "Cannot instantiate a serverHTTP handler")
			}
		})
	}
}
