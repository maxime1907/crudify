package benchmark

import (
	"net/http"
	"strconv"
	"time"
)

func testDelete(quantity int) (time.Duration, error) {
	start := time.Now()
	for i := 1; i <= quantity; i++ {
		myurlReq := url + "crudify" + "?id=-" + strconv.Itoa(i)
		req, err := http.NewRequest("DELETE", myurlReq, nil)
		if err != nil {
			return time.Since(start), err
		}
		err = execRequest(req)
		if err != nil {
			return time.Since(start), err
		}
	}
	return time.Since(start), nil
}
