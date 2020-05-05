package test

import (
	"net/http"
	"testing"
)

func testDelete(t *testing.T) {
	req, err := http.NewRequest("DELETE", url+"crudify?id=-1", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = execRequest(req)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func testDeleteAllTest(t *testing.T) {
	req, err := http.NewRequest("DELETE", url+"crudify?id=-1", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = execRequest(req)
	if err != nil {
		t.Fatal(err.Error())
	}
	req, err = http.NewRequest("DELETE", url+"crudify?id=-2", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = execRequest(req)
	if err != nil {
		t.Fatal(err.Error())
	}
	req, err = http.NewRequest("DELETE", url+"crudify?id=-3", nil)
	if err != nil {
		t.Fatal(err.Error())
	}
	err = execRequest(req)
	if err != nil {
		t.Fatal(err.Error())
	}
}
