package router

import (
	"net/http"

	"github.com/maxime1907/crudify/handler"
)

type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
}

var myroutes = []Route{
	//TEST
	Route{"test_get", "GET", "/test", handler.TestGet},
}
