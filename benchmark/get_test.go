package benchmark

import (
	"net/http"
	"strconv"
	"time"
)

func testGet(quantity int) (time.Duration, error) {
	start := time.Now()
	myurlReq := url + "crudify" + "?_limit=" + strconv.Itoa(quantity)
	req, err := http.NewRequest("GET", myurlReq, nil)
	if err != nil {
		return time.Since(start), err
	}
	err = execRequest(req)
	if err != nil {
		return time.Since(start), err
	}
	return time.Since(start), nil
}
