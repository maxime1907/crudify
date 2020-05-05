package test

import (
	"net/http"
	"testing"
)

func testRoot(t *testing.T) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		t.Fatal(err)
	}
	err = execRequest(req)
	if err != nil {
		t.Fatal(err)
	}
}

func testGetSingle(t *testing.T) {
	req, err := http.NewRequest("GET", url+"crudify?id=0", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = execRequest(req)
	if err != nil {
		t.Fatal(err)
	}
}

func testGetMultiple(t *testing.T) {
	req, err := http.NewRequest("GET", url+"crudify", nil)
	if err != nil {
		t.Fatal(err)
	}
	err = execRequest(req)
	if err != nil {
		t.Fatal(err)
	}
}
