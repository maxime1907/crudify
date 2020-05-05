package benchmark

import (
	"bytes"
	"net/http"
	"strconv"
	"time"
)

func testPut(quantity int) (time.Duration, error) {
	var jsonStr = []byte(`[`)
	for i := 1; i <= quantity; i++ {
		nb := strconv.Itoa(i)
		jsonStr = append(jsonStr, []byte(`
			{
				"id" : "-`+nb+`",
				"name" : "testPostSingleIssou`+nb+`",
				"creation" : "2017-02-10",
				"description" : "salut tout le monde",
				"admin" : true
			}
		`)...)
		if i+1 <= quantity {
			jsonStr = append(jsonStr, []byte(`,`)...)
		}
	}
	jsonStr = append(jsonStr, []byte(`]`)...)

	start := time.Now()
	req, err := http.NewRequest("PUT", url+"crudify", bytes.NewBuffer(jsonStr))
	if err != nil {
		return time.Since(start), err
	}
	return time.Since(start), execRequest(req)
}
