package handler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestGetStatusCode(t *testing.T) {
	type args struct {
		er error
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "OK status code",
			args: args{
				er: nil,
			},
			want: http.StatusOK,
		},
		{
			name: "NotFound status code",
			args: args{
				er: errors.New("sql: no rows in result set"),
			},
			want: http.StatusNotFound,
		},
		{
			name: "Conflict status code",
			args: args{
				er: errors.New("pq: duplicate key value violates unique constraint"),
			},
			want: http.StatusConflict,
		},
		{
			name: "InternalServerError status code",
			args: args{
				er: errors.New("strconv.Atoi: parsing \"memelord\": invalid syntax"),
			},
			want: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetStatusCode(nil, tt.args.er); got != tt.want {
				t.Errorf("GetStatusCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSendAnswer(t *testing.T) {
	type args struct {
		w    http.ResponseWriter
		data interface{}
		er   error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Normal response",
			args: args{
				w:    httptest.NewRecorder(),
				data: nil,
				er:   nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SendAnswer(tt.args.w, nil, tt.args.data, tt.args.er)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendAnswer() error = %v", err)
				return
			}
		})
	}
}

func TestFormToMap(t *testing.T) {
	req, err := http.NewRequest("GET", "/test?parameter1=toto&parameter2=hola", nil)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "Valid conversion",
			args: args{
				r: req,
			},
			want: map[string]string{"parameter1": "toto", "parameter2": "hola"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormToMap(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FormToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetTableName(t *testing.T) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatal(err)
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Valid url on GET",
			args: args{
				r: req,
			},
			want: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetTableName(tt.args.r); got != tt.want {
				t.Errorf("getTableName() = %v, want %v", got, tt.want)
			}
		})
	}
}
