package test

import (
	"bytes"
	"net/http"
	"testing"
)

func testPutSingle(t *testing.T) {
	var jsonStr = []byte(`
		{
			"id" : "-1",
			"name" : "testPostSingleIssou",
			"creation" : "2017-02-10",
			"description" : "salut tout le monde",
			"admin" : true
		}
	`)
	req, err := http.NewRequest("PUT", url+"crudify", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	err = execRequest(req)
	if err != nil {
		t.Fatal(err)
	}
}

func testPutMultiple(t *testing.T) {
	var jsonStr = []byte(`
		[
		{
			"id" : "-2",
			"name" : "testPostSingleIssou1",
			"creation" : "2017-02-10",
			"description" : "salut tout le monde",
			"admin" : true
		},
		{
			"id" : "-3",
			"name" : "testPostSingleIssou2",
			"creation" : "2017-02-10",
			"description" : "salut tout le monde",
			"admin" : true
		}
		]
	`)
	req, err := http.NewRequest("PUT", url+"crudify", bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	err = execRequest(req)
	if err != nil {
		t.Fatal(err)
	}
}
